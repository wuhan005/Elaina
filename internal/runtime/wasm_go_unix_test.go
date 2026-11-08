//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
)

func TestWASMGo(t *testing.T) {
	dir := os.Getenv("WASM_TEST_DIR")
	if dir == "" {
		t.Skip("set WASM_TEST_DIR to run Go WASM integration tests")
	}
	if _, err := os.Stat(filepath.Join(dir, "go", "importcfg")); err != nil {
		t.Fatal("run scripts/build-wasm-go.sh before the Go WASM tests: ", err)
	}

	original := config.App
	t.Cleanup(func() { config.App = original })
	config.App.WASMRuntimeDir = dir
	config.App.WASMCacheDir = t.TempDir()
	if cache := os.Getenv("WASM_TEST_CACHE"); cache != "" {
		config.App.WASMCacheDir = cache
	}
	if err := Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(context.Background()) })

	run := func(t *testing.T, code string, internet bool, memory int64, timeout int) *ExecOutput {
		t.Helper()

		task, err := NewWASMTask(context.Background(), NewWASMTaskOptions{
			Language: "go",
			Template: &db.Tpl{
				MaxMemory:      memory,
				Timeout:        timeout,
				InternetAccess: internet,
			},
			Code: []byte(code),
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

	t.Run("compile and run", func(t *testing.T) {
		t.Setenv("ELAINA_HOST_SECRET", "must-not-leak")
		output := run(t, `package main
import ("fmt"; "os"; "time")

func main() {
 if os.Getenv("ELAINA_HOST_SECRET") != "" { panic("host environment exposed") }
 if _, err := os.ReadFile("/toolchain/importcfg"); err == nil { panic("toolchain exposed") }
 if err := os.WriteFile("result.txt", []byte("go-wasm-ok"), 0600); err != nil { panic(err) }
 result := make(chan string)
 go func() { time.Sleep(time.Millisecond); data, err := os.ReadFile("result.txt"); if err != nil { panic(err) }; result <- string(data) }()
 fmt.Println(<-result)
}`, false, 512, 60)
		if output.Error != "" || strings.TrimSpace(output.Stdout) != "go-wasm-ok" {
			t.Fatalf("Go WASM compile/link/run: %+v", output)
		}
	})

	t.Run("HTTP and DNS", func(t *testing.T) {
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			_, _ = w.Write([]byte("network-ok"))
		}))
		defer server.Close()

		for _, address := range []string{server.URL, strings.Replace(server.URL, "127.0.0.1", "localhost", 1)} {
			code := fmt.Sprintf(`package main
import ("fmt"; "io"; "net/http"; "os")
func main() { r, err := http.Get(%q); if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }; defer r.Body.Close(); _, _ = io.Copy(os.Stdout, r.Body) }`, address)
			before := requests.Load()
			output := run(t, code, true, 512, 60)
			if output.Error != "" || output.Stdout != "network-ok" || requests.Load() != before+1 {
				t.Fatalf("network enabled: %+v requests=%d", output, requests.Load())
			}

			output = run(t, code, false, 512, 60)
			if output.Error == "" || requests.Load() != before+1 {
				t.Fatalf("network disabled: %+v requests=%d", output, requests.Load())
			}
		}
	})

	t.Run("HTTPS", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("tls-ok"))
		}))
		defer server.Close()

		cert := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: server.Certificate().Raw,
		})
		certificates, err := wasmGoRootCertificates()
		if err != nil {
			t.Fatal(err)
		}

		code := fmt.Sprintf(`package main
import ("crypto/tls"; "crypto/x509"; "fmt"; "io"; "net/http"; "os")
func main() {
 roots, err := x509.SystemCertPool(); if err != nil { panic(err) }
 if %t && len(roots.Subjects()) == 0 { panic("system roots missing") }
 if r, err := http.Get(%q); err == nil { r.Body.Close(); panic("untrusted certificate accepted") }
 roots.AppendCertsFromPEM([]byte(%q))
 http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs:roots}
 r, err := http.Get(%q); if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }; defer r.Body.Close(); _, _ = io.Copy(os.Stdout, r.Body)
}`, len(certificates) > 0, server.URL, string(cert), server.URL)
		output := run(t, code, true, 512, 60)
		if output.Error != "" || output.Stdout != "tls-ok" {
			t.Fatalf("HTTPS with verified certificate: %+v", output)
		}
	})

	t.Run("errors and limits", func(t *testing.T) {
		for _, test := range []struct {
			name, code, error, stderr string
			memory                    int64
			timeout                   int
		}{
			{"compiler diagnostic", "package main; func main() { missing() }", "exit status", "undefined: missing", 512, 60},
			{"unavailable import", "package main; import _ \"example.com/missing\"; func main() {}", "not available", "", 512, 60},
			{"exit status", "package main; import \"os\"; func main() { os.Exit(7) }", "exit status 7", "", 512, 60},
			{"memory", "package main; func main() {}", "", "", 8, 60},
			{"timeout", "package main; func main() { for {} }", "execution timed out", "", 512, 10},
			{"output", "package main; import \"fmt\"; func main() { for { fmt.Println(\"output-limit\") } }", "output limit exceeded", "", 512, 60},
		} {
			t.Run(test.name, func(t *testing.T) {
				output := run(t, test.code, false, test.memory, test.timeout)
				if output.Error == "" || !strings.Contains(output.Error, test.error) || !strings.Contains(output.Stderr, test.stderr) {
					t.Fatalf("Go failure: %+v", output)
				}
			})
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

		code := fmt.Sprintf("package main; import \"net/http\"; func main() { _, _ = http.Get(%q) }", server.URL)
		task, err := NewWASMTask(ctx, NewWASMTaskOptions{
			Language: "go",
			Template: &db.Tpl{
				MaxMemory:      512,
				Timeout:        60,
				InternetAccess: true,
			},
			Code: []byte(code),
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = task.Run(ctx)
		if err == nil || !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("request cancellation: %v", err)
		}

		select {
		case start := <-entered:
			if time.Since(start) > 2*time.Second {
				t.Fatal("blocked HTTP did not cancel promptly")
			}
		default:
			t.Fatal("Go program did not reach the HTTP server")
		}
	})
}
