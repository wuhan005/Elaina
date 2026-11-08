//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
)

func TestWASMC(t *testing.T) {
	dir := os.Getenv("WASM_TEST_DIR")
	if dir == "" {
		t.Skip("set WASM_TEST_DIR to run C WASM integration tests")
	}
	original := config.App
	t.Cleanup(func() { config.App = original })
	config.App.WASMRuntimeDir = dir
	config.App.WASMCacheDir = t.TempDir()
	if cache := os.Getenv("WASM_TEST_CACHE"); cache != "" {
		config.App.WASMCacheDir = cache
	}
	cache, err := wazero.NewCompilationCacheWithDir(config.App.WASMCacheDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cache.Close(context.Background()) })
	run := func(t *testing.T, code string, internet bool, timeout int) *ExecOutput {
		t.Helper()
		task, err := NewWASMTask(context.Background(), NewWASMTaskOptions{
			Language: "c", Template: &db.Tpl{MaxMemory: 512, Timeout: timeout, InternetAccess: internet}, Code: []byte(code),
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
	compile := func(t *testing.T, ctx context.Context, workspace, code string, memory int64) (*wasmOutputBuffer, error) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(workspace, "main.c"), []byte(code), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		output := &wasmOutputBuffer{limit: wasmOutputLimit, cancel: cancel}
		task := &WASMTask{template: &db.Tpl{MaxMemory: memory, InternetAccess: true}}
		err := task.compileCWASM(ctx, cache, filepath.Join(dir, "c"), workspace, output)
		return output, err
	}
	t.Run("compile and run", func(t *testing.T) {
		t.Setenv("ELAINA_HOST_SECRET", "must-not-leak")
		output := run(t, `#include <stdio.h>
#include <stdlib.h>
#include <string.h>
int main(void) {
 if (getenv("ELAINA_HOST_SECRET")) return 1;
 if (fopen("/toolchain/network.c", "r")) return 2;
 FILE *file = fopen("result.txt", "w+"); if (!file) return 3;
 fputs("c-wasm-ok", file); rewind(file);
 char buffer[32]; if (!fgets(buffer, sizeof(buffer), file)) return 4;
 fclose(file); puts(buffer); return strcmp(buffer, "c-wasm-ok");
}`, false, 60)
		if output.Error != "" || strings.TrimSpace(output.Stdout) != "c-wasm-ok" {
			t.Fatalf("C WASM compile/link/run: %+v", output)
		}
	})
	t.Run("HTTP and DNS", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			_, _ = w.Write([]byte("network-ok"))
		}))
		defer server.Close()
		port := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")[1]
		for _, host := range []string{"127.0.0.1", "localhost"} {
			code := wasmCHTTPClient(host, port)
			before := requests.Load()
			output := run(t, code, true, 60)
			if output.Error != "" || !strings.Contains(output.Stdout, "network-ok") || requests.Load() != before+1 {
				t.Fatalf("network enabled: %+v requests=%d", output, requests.Load())
			}
			output = run(t, code, false, 60)
			if output.Error == "" || requests.Load() != before+1 {
				t.Fatalf("network disabled: %+v requests=%d", output, requests.Load())
			}
		}
	})
	t.Run("blocked HTTP cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		entered := make(chan time.Time, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entered <- time.Now()
			cancel()
			<-r.Context().Done()
		}))
		defer server.Close()
		port := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")[1]
		task, err := NewWASMTask(ctx, NewWASMTaskOptions{
			Language: "c", Template: &db.Tpl{MaxMemory: 512, Timeout: 60, InternetAccess: true},
			Code: []byte(wasmCHTTPClient("127.0.0.1", port)),
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = task.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("request cancellation: %v", err)
		}
		select {
		case start := <-entered:
			if time.Since(start) > 2*time.Second {
				t.Fatal("blocked C socket did not cancel promptly")
			}
		default:
			t.Fatal("C program did not reach the HTTP server")
		}
	})
	t.Run("compiler cannot read host files", func(t *testing.T) {
		parent := t.TempDir()
		secret := "ELAINA_PRIVATE_COMPILE_SENTINEL_73b924"
		path := filepath.Join(parent, "secret.h")
		if err := os.WriteFile(path, []byte("#error "+secret+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		for _, code := range []string{
			fmt.Sprintf("#include %q\nint main(void) {return 0;}", path),
			"#include \"../secret.h\"\nint main(void) {return 0;}",
			fmt.Sprintf("const char data[] = {\n#embed %q\n};\nint main(void) {return 0;}", path),
		} {
			workspace, err := os.MkdirTemp(parent, "compile-")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			output, err := compile(t, ctx, workspace, code, 512)
			cancel()
			diagnostic := output.buffer.String()
			if err == nil || strings.Contains(diagnostic, secret) || (!strings.Contains(diagnostic, "file not found") && !strings.Contains(diagnostic, "Operation not permitted")) {
				t.Fatalf("compiler file isolation: %v %s", err, output.buffer.String())
			}
		}
	})
	t.Run("compiler cannot execute commands", func(t *testing.T) {
		workspace := t.TempDir()
		marker := filepath.Join(t.TempDir(), "command-executed")
		command := "touch " + marker
		if err := os.WriteFile(filepath.Join(workspace, "Makefile"), []byte("all:\n\t"+command+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		code := fmt.Sprintf(`__attribute__((import_module("env"), import_name("system"))) extern int command(const char *);
__attribute__((constructor)) static void initialize(void) { command(%q); }
int main(void) { return 0; }`, command)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if output, err := compile(t, ctx, workspace, code, 512); err != nil {
			t.Fatalf("compile command fixture: %v %s", err, output.buffer.String())
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("compilation executed a command")
		}
		output := &wasmOutputBuffer{limit: wasmOutputLimit, cancel: cancel}
		task := &WASMTask{template: &db.Tpl{MaxMemory: 512}}
		err := task.runWASMModule(ctx, nil, wasmModuleOptions{
			path: filepath.Join(workspace, "main.wasm"), args: []string{"main"}, dir: workspace,
		}, output, output)
		if err == nil || !strings.Contains(err.Error(), "env") {
			t.Fatalf("host command import was not rejected: %v", err)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("WASM execution escaped to a host command")
		}
		compiler := t.TempDir()
		probe, err := os.ReadFile(filepath.Join(workspace, "main.wasm"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(compiler, "clang.wasm"), probe, 0600); err != nil {
			t.Fatal(err)
		}
		if err := task.compileCWASM(ctx, cache, compiler, workspace, output); err == nil || !strings.Contains(err.Error(), "env") {
			t.Fatalf("compiler host command import was not rejected: %v", err)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("compiler probe executed a host command")
		}
	})
	t.Run("compiler network is always denied", func(t *testing.T) {
		workspace := t.TempDir()
		code := `#include <errno.h>
#include <netdb.h>
#include <stdio.h>
int main(void) {
 struct addrinfo hints = {.ai_family = AF_INET, .ai_socktype = SOCK_STREAM}, *result;
 if (getaddrinfo("localhost", "80", &hints, &result) != EAI_SYSTEM || errno != EPERM) return 1;
 if (socket(AF_INET, SOCK_STREAM, 0) != -1 || errno != EPERM) return 2;
 puts("compiler-network-denied"); return 42;
}`
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if output, err := compile(t, ctx, workspace, code, 512); err != nil {
			t.Fatalf("compile network probe: %v %s", err, output.buffer.String())
		}
		probe, err := os.ReadFile(filepath.Join(workspace, "main.wasm"))
		if err != nil {
			t.Fatal(err)
		}
		compiler := t.TempDir()
		if err := os.WriteFile(filepath.Join(compiler, "clang.wasm"), probe, 0600); err != nil {
			t.Fatal(err)
		}
		output := &wasmOutputBuffer{limit: wasmOutputLimit, cancel: cancel}
		task := &WASMTask{template: &db.Tpl{MaxMemory: 512, InternetAccess: true}}
		err = task.compileCWASM(ctx, cache, compiler, workspace, output)
		if err == nil || !strings.Contains(output.buffer.String(), "compiler-network-denied") {
			t.Fatalf("compiler network gate: %v %s", err, output.buffer.String())
		}
	})
	t.Run("compiler resource limits", func(t *testing.T) {
		large := "#define M0 1\n"
		for i := 1; i <= 26; i++ {
			large += fmt.Sprintf("#define M%d M%d,M%d\n", i, i-1, i-1)
		}
		large += "int data[] = {M26}; int main(void) {return 0;}"
		for _, test := range []struct {
			name, code string
			memory     int64
			timeout    time.Duration
		}{
			{"memory", large, 128, 30 * time.Second},
			{"timeout", large, 512, time.Second},
			{"output", strings.Repeat("#warning "+strings.Repeat("x", 2048)+"\n", 1024) + "int main(void) {return 0;}", 512, 30 * time.Second},
		} {
			t.Run(test.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
				defer cancel()
				start := time.Now()
				output, err := compile(t, ctx, t.TempDir(), test.code, test.memory)
				if err == nil {
					t.Fatal("compiler exceeded its resource limit")
				}
				switch test.name {
				case "memory":
					if ctx.Err() != nil {
						t.Fatalf("memory failure only stopped at deadline: %v", err)
					}
					if !strings.Contains(strings.ToLower(err.Error()+output.buffer.String()), "memory") {
						t.Fatalf("unexpected compiler failure: %v %s", err, output.buffer.String())
					}
				case "timeout":
					if !errors.Is(ctx.Err(), context.DeadlineExceeded) || time.Since(start) > 5*time.Second {
						t.Fatalf("compiler timeout: %v after %s", err, time.Since(start))
					}
				case "output":
					if !output.exceeded || output.buffer.Len() != wasmOutputLimit {
						t.Fatalf("compiler output limit: %v bytes=%d", err, output.buffer.Len())
					}
				}
			})
		}
	})
	t.Run("execution limits and errors", func(t *testing.T) {
		for _, test := range []struct {
			name, code, error string
			timeout           int
		}{
			{"exit", "int main(void) {return 7;}", "exit status 7", 60},
			{"loop", "int main(void) {for (;;) {}}", "execution timed out", 3},
			{"output", "#include <stdio.h>\nint main(void) {for (;;) puts(\"output-limit\");}", "output limit exceeded", 60},
			{"invalid source", "int main(void) {missing();}", "exit status", 60},
		} {
			t.Run(test.name, func(t *testing.T) {
				output := run(t, test.code, false, test.timeout)
				if !strings.Contains(output.Error, test.error) {
					t.Fatalf("C error: %+v", output)
				}
			})
		}
		output := run(t, "#include <stdlib.h>\n#include <stdio.h>\nint main(void) {void *p=malloc(768u*1024*1024); if(p) {free(p); return 1;} puts(\"memory-limited\");}", false, 60)
		if output.Error != "" || strings.TrimSpace(output.Stdout) != "memory-limited" {
			t.Fatalf("runtime memory limit: %+v", output)
		}
	})
}

func wasmCHTTPClient(host, port string) string {
	return fmt.Sprintf(`#include <stdio.h>
#include <netdb.h>
#include <unistd.h>
int main(void) {
 struct addrinfo hints = {.ai_family = AF_INET, .ai_socktype = SOCK_STREAM}, *address;
 if (getaddrinfo(%q, %q, &hints, &address)) { perror("getaddrinfo"); return 1; }
 int fd = socket(address->ai_family, address->ai_socktype, address->ai_protocol);
 if (fd < 0 || connect(fd, address->ai_addr, address->ai_addrlen)) { perror("connect"); return 2; }
 freeaddrinfo(address);
 const char request[] = "GET / HTTP/1.0\r\nHost: localhost\r\n\r\n";
 if (send(fd, request, sizeof(request)-1, 0) != sizeof(request)-1) return 3;
 char buffer[256]; int count;
 while ((count = recv(fd, buffer, sizeof(buffer), 0)) > 0) fwrite(buffer, 1, count, stdout);
 close(fd); return count < 0;
}`, host, port)
}
