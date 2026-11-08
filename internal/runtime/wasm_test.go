// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
)

func TestNewTask(t *testing.T) {
	for _, language := range []string{"php", "python", "javascript", "go", "c"} {
		task, err := NewTask(context.Background(), NewTaskOptions{
			Language: language,
			Template: &db.Tpl{
				MaxMemory:      128,
				Timeout:        10,
				InternetAccess: true,
			},
		})
		if err != nil {
			t.Fatalf("%s: %v", language, err)
		}
		if _, ok := task.(*WASMTask); !ok {
			t.Fatalf("%s: got %T, want WASM", language, task)
		}
	}

	if _, err := NewTask(context.Background(), NewTaskOptions{
		Language: "unknown",
		Template: &db.Tpl{},
	}); err == nil {
		t.Fatal("expected an error for an unsupported language")
	}
}

func TestNewWASMTaskValidation(t *testing.T) {
	for _, template := range []*db.Tpl{
		nil,
		{MaxMemory: -1},
		{MaxMemory: 4097},
		{MaxCPUs: -1},
		{MaxCPUs: 11},
		{MaxCPUs: 1},
		{Timeout: -1},
		{Timeout: 61},
	} {
		if _, err := NewWASMTask(context.Background(), NewWASMTaskOptions{
			Language: "python",
			Template: template,
		}); err == nil {
			t.Fatalf("expected an error for template %+v", template)
		}
	}

	if _, err := NewWASMTask(context.Background(), NewWASMTaskOptions{
		Language: "python",
		Template: &db.Tpl{
			InternetAccess: true,
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWASMOutputLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buffer := &wasmOutputBuffer{
		limit:  wasmOutputLimit,
		cancel: cancel,
	}
	if n, err := buffer.Write(bytes.Repeat([]byte{'x'}, wasmOutputLimit)); n != wasmOutputLimit || err != nil {
		t.Fatalf("write at limit: %d, %v", n, err)
	}

	if n, err := buffer.Write([]byte("overflow")); n != 0 || err == nil {
		t.Fatalf("write beyond limit: %d, %v", n, err)
	}
	if !buffer.exceeded || buffer.buffer.Len() != wasmOutputLimit || ctx.Err() == nil {
		t.Fatal("output overflow did not cancel execution with bounded output")
	}

	buffer = &wasmOutputBuffer{
		limit:  wasmOutputLimit,
		cancel: cancel,
	}
	reader := struct{ io.Reader }{
		bytes.NewReader(bytes.Repeat([]byte{'x'}, 2*wasmOutputLimit)),
	}
	if _, err := io.Copy(buffer, reader); err == nil || buffer.buffer.Len() != wasmOutputLimit {
		t.Fatal("io.Copy bypassed the output limit")
	}
}

func TestWASMInterpreters(t *testing.T) {
	dir := os.Getenv("WASM_TEST_DIR")
	if dir == "" {
		t.Skip("set WASM_TEST_DIR to run the interpreter integration tests")
	}

	original := config.App
	t.Cleanup(func() { config.App = original })
	config.App.WASMRuntimeDir = dir
	config.App.WASMCacheDir = t.TempDir()
	if cache := os.Getenv("WASM_TEST_CACHE"); cache != "" {
		config.App.WASMCacheDir = cache
	}

	t.Setenv("ELAINA_TEST_SECRET", "host-only")

	run := func(t *testing.T, language, code string, internet bool, memory int64, timeout int) *ExecOutput {
		t.Helper()

		task, err := NewWASMTask(context.Background(), NewWASMTaskOptions{
			Language: language,
			Code:     []byte(code),
			Template: &db.Tpl{
				InternetAccess: internet,
				MaxMemory:      memory,
				Timeout:        timeout,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		output, err := task.Run(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		return output
	}

	for _, test := range []struct{ language, code, exit, loop, sleep string }{
		{
			"python",
			`import os, sys, time
assert os.getenv('ELAINA_TEST_SECRET') is None
assert not os.path.exists('/etc/passwd')
assert not os.path.exists('data.txt')
assert sys.stdin.read() == ''
try:
    os.symlink('/etc/passwd', 'host-link')
    raise AssertionError('symlinks must be denied')
except OSError: pass
with open('data.txt', 'w') as f: f.write('file-ok')
time.sleep(0.001)
print(open('data.txt').read())
sys.stderr.write('stderr-ok')`,
			`raise SystemExit(7)`,
			"print('started', flush=True)\nwhile True: pass",
			"import time\nprint('started', flush=True)\ntime.sleep(60)",
		},
		{
			"php",
			`<?php
if (getenv('ELAINA_TEST_SECRET') || file_exists('/etc/passwd') || file_exists('data.txt')) exit(1);
if (stream_get_contents(STDIN) !== '') exit(2);
file_put_contents('data.txt', 'file-ok');
usleep(1000);
echo file_get_contents('data.txt'); fwrite(STDERR, 'stderr-ok');`,
			`<?php exit(7);`,
			`<?php echo 'started'; while (true) {}`,
			`<?php echo 'started'; sleep(60);`,
		},
		{
			"javascript",
			`import * as std from 'qjs:std';
import * as os from 'qjs:os';
if (std.getenv('ELAINA_TEST_SECRET') || std.loadFile('/etc/passwd') || std.loadFile('data.txt')) std.exit(1);
const f = std.open('data.txt', 'w'); f.puts('file-ok'); f.close();
os.setTimeout(() => { console.log(std.loadFile('data.txt')); std.err.puts('stderr-ok'); }, 10);`,
			`import * as std from 'qjs:std'; std.exit(7);`,
			`console.log('started'); while (true) {}`,
			`import * as os from 'qjs:os'; console.log('started'); os.setTimeout(() => {}, 60000);`,
		},
	} {
		t.Run(test.language, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				output := run(t, test.language, test.code, false, 128, 60)
				if output.Error != "" || strings.TrimSpace(output.Stdout) != "file-ok" || strings.TrimSpace(output.Stderr) != "stderr-ok" {
					t.Fatalf("filesystem/stdio/isolation: %+v", output)
				}
			}

			if output := run(t, test.language, test.exit, false, 128, 60); output.Error != "exit status 7" {
				t.Fatalf("exit: %+v", output)
			}

			for _, code := range []string{test.loop, test.sleep} {
				start := time.Now()
				if output := run(t, test.language, code, false, 128, 10); output.Error != "execution timed out" || strings.TrimSpace(output.Stdout) != "started" || time.Since(start) > 12*time.Second {
					t.Fatalf("timeout: %+v after %s", output, time.Since(start))
				}
			}
		})
	}

	t.Run("PHP networking", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			_, _ = w.Write([]byte("network-ok"))
		}))
		defer server.Close()

		code := fmt.Sprintf(`<?php $r = file_get_contents(%q); if ($r === false) exit(1); echo $r;`, server.URL)
		output := run(t, "php", code, true, 128, 60)
		if output.Error != "" || output.Stdout != "network-ok" || requests.Load() != 1 {
			t.Fatalf("network enabled: %+v, requests=%d", output, requests.Load())
		}

		output = run(t, "php", code, false, 128, 60)
		if output.Error == "" || strings.Contains(output.Stdout, "network-ok") || requests.Load() != 1 {
			t.Fatalf("network disabled: %+v, requests=%d", output, requests.Load())
		}

		code = strings.Replace(code, "127.0.0.1", "localhost", 1)
		output = run(t, "php", code, true, 128, 60)
		if output.Error == "" || requests.Load() != 1 {
			t.Fatalf("DNS compatibility changed; update the documented restriction: %+v", output)
		}

		output = run(t, "php", `<?php if (in_array('https', stream_get_wrappers()) || function_exists('curl_init')) exit(1); echo 'no-tls';`, true, 128, 60)
		if output.Error != "" || output.Stdout != "no-tls" {
			t.Fatalf("HTTPS compatibility changed: %+v", output)
		}
	})

	t.Run("blocked network timeout", func(t *testing.T) {
		if output := run(t, "php", `<?php echo 'ready';`, false, 128, 60); output.Error != "" {
			t.Fatalf("warm PHP compilation cache: %+v", output)
		}

		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			<-r.Context().Done()
		}))
		defer server.Close()

		code := fmt.Sprintf(`<?php echo 'started'; file_get_contents(%q);`, server.URL)
		start := time.Now()
		output := run(t, "php", code, true, 128, 10)
		if output.Error != "execution timed out" || output.Stdout != "started" || requests.Load() != 1 || time.Since(start) > 12*time.Second {
			t.Fatalf("blocking host call survived timeout: %+v requests=%d", output, requests.Load())
		}
	})

	t.Run("blocked network cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		connected := make(chan struct{})
		disconnected := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(connected)
			<-r.Context().Done()
			close(disconnected)
		}))
		defer func() {
			cancel()
			server.Close()
		}()

		task, err := NewWASMTask(ctx, NewWASMTaskOptions{
			Language: "php",
			Template: &db.Tpl{
				Timeout:        60,
				MaxMemory:      128,
				InternetAccess: true,
			},
			Code: []byte(fmt.Sprintf(`<?php file_get_contents(%q);`, server.URL)),
		})
		if err != nil {
			t.Fatal(err)
		}

		done := make(chan error, 1)
		go func() { _, err := task.Run(ctx); done <- err }()

		select {
		case <-connected:
		case err := <-done:
			t.Fatalf("task exited before connecting: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("task did not connect")
		}

		cancel()

		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("request cancellation: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("task did not stop after cancellation")
		}

		select {
		case <-disconnected:
		case <-time.After(time.Second):
			t.Fatal("cancelled task left its socket open")
		}
	})

	t.Run("documented limitations", func(t *testing.T) {
		python := `import socket, time
try:
    socket.getaddrinfo('localhost', 80)
    raise AssertionError('DNS support changed')
except AttributeError: pass
try:
    socket.socket()
    raise AssertionError('socket support changed')
except OSError: pass
try:
    import ssl
    raise AssertionError('TLS support changed')
except ModuleNotFoundError: pass
time.sleep(0.001)
print('limitations-ok')`
		if output := run(t, "python", python, true, 128, 60); output.Error != "" || strings.TrimSpace(output.Stdout) != "limitations-ok" {
			t.Fatalf("Python compatibility: %+v", output)
		}

		javascript := `if (typeof require !== 'undefined' || typeof fetch !== 'undefined' || typeof setTimeout !== 'undefined') throw Error('capabilities changed'); console.log('limitations-ok');`
		if output := run(t, "javascript", javascript, true, 128, 60); output.Error != "" || strings.TrimSpace(output.Stdout) != "limitations-ok" {
			t.Fatalf("JavaScript compatibility: %+v", output)
		}
	})

	t.Run("concurrent tasks", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				marker := fmt.Sprintf("task-%d", i)
				code := fmt.Sprintf("import sys\nprint(%q)\nsys.stderr.write(%q)", marker, marker)
				if output := run(t, "python", code, false, 128, 60); output.Error != "" || strings.TrimSpace(output.Stdout) != marker || output.Stderr != marker {
					t.Errorf("concurrent task: %+v", output)
				}
			}()
		}

		wg.Wait()
	})

	if output := run(t, "python", `print('x' * (2 * 1024 * 1024))`, false, 128, 60); output.Error != "output limit exceeded" || len(output.Stdout) > wasmOutputLimit {
		t.Fatalf("output limit: error=%s bytes=%d", output.Error, len(output.Stdout))
	}

	if output := run(t, "python", `data = bytearray(128 * 1024 * 1024)`, false, 64, 60); output.Error == "" || !strings.Contains(output.Stderr, "MemoryError") {
		t.Fatalf("linear memory limit: %+v", output)
	}

	t.Run("request cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		task, _ := NewWASMTask(ctx, NewWASMTaskOptions{
			Language: "python",
			Template: &db.Tpl{
				Timeout: 60,
			},
			Code: []byte("while True: pass"),
		})
		if _, err := task.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("got %v, want request cancellation", err)
		}
	})
}
