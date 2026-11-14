// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flamego/flamego"

	"github.com/wuhan005/Elaina/internal/config"
	appcontext "github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/runtime"
)

func TestRequestBodyLimit(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxRequestBytes = 32
	for _, test := range []struct {
		name, body string
		unknown    bool
		status     int
	}{
		{"exact", strings.Repeat("x", 32), false, http.StatusOK},
		{"known oversized", strings.Repeat("x", 33), false, http.StatusRequestEntityTooLarge},
		{"chunked exact", strings.Repeat("x", 32), true, http.StatusOK},
		{"chunked oversized", strings.Repeat("x", 33), true, http.StatusRequestEntityTooLarge},
		{"chunked trailing JSON", "{}" + strings.Repeat(" ", 31), true, http.StatusRequestEntityTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			called := false
			f := flamego.New()
			f.Use(limitRequestBody)
			f.Post("/", func(ctx flamego.Context) {
				called = true
				ctx.ResponseWriter().WriteHeader(http.StatusOK)
			})
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			if test.unknown {
				r.ContentLength = -1
				r.TransferEncoding = []string{"chunked"}
			}
			w := httptest.NewRecorder()
			f.ServeHTTP(w, r)
			if w.Code != test.status || called != (test.status == http.StatusOK) {
				t.Fatalf("status=%d called=%v body=%s", w.Code, called, w.Body.String())
			}
			if test.status != http.StatusOK {
				var response dto.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Msg == "" {
					t.Fatalf("invalid error envelope: %s, %v", w.Body.String(), err)
				}
			}
		})
	}
}

func TestRequestBodyBudget(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxRequestBytes = 32
	config.App.WASMLimits.MaxBufferedRequestBytes = 48

	entered, unblock := make(chan struct{}), make(chan struct{})
	f := flamego.New()
	f.Use(limitRequestBody)
	f.Post("/", func(ctx flamego.Context) {
		if ctx.Request().Header.Get("X-Block") != "" {
			entered <- struct{}{}
			<-unblock
		}
		ctx.ResponseWriter().WriteHeader(http.StatusOK)
	})
	send := func(size int, chunked, block bool) int {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", size)))
		if chunked {
			r.ContentLength = -1
			r.TransferEncoding = []string{"chunked"}
		}
		if block {
			r.Header.Set("X-Block", "1")
		}
		w := httptest.NewRecorder()
		f.ServeHTTP(w, r)
		return w.Code
	}

	done := make(chan int)
	go func() { done <- send(32, false, true) }()
	<-entered
	for _, test := range []struct {
		size    int
		chunked bool
		status  int
	}{
		{17, false, http.StatusServiceUnavailable},
		{16, false, http.StatusOK},
		{1, true, http.StatusServiceUnavailable},
		{0, false, http.StatusOK},
	} {
		if status := send(test.size, test.chunked, false); status != test.status {
			t.Fatalf("size=%d chunked=%v: status=%d, want %d", test.size, test.chunked, status, test.status)
		}
	}
	close(unblock)
	if status := <-done; status != http.StatusOK {
		t.Fatalf("blocked request: %d", status)
	}
	if used := bufferedRequestBytes.Load(); used != 0 {
		t.Fatalf("request memory budget leaked %d bytes", used)
	}
	if status := send(32, true, false); status != http.StatusOK {
		t.Fatalf("budget was not released: %d", status)
	}
}

func TestClientAddress(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.TrustedProxies = []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	for _, test := range []struct{ remote, forwarded, want string }{
		{"203.0.113.7:1234", "198.51.100.9", "203.0.113.7"},
		{"10.0.0.2:1234", "", "10.0.0.2"},
		{"10.0.0.2:1234", "1.2.3.4, 198.51.100.9", "198.51.100.9"},
		{"10.0.0.2:1234", "198.51.100.9, 10.0.0.3", "198.51.100.9"},
		{"10.0.0.2:1234", "invalid, 10.0.0.3", "10.0.0.3"},
		{"[2001:db8:1:2:3:4:5:6]:1234", "", "2001:db8:1:2::/64"},
		{"[::ffff:192.0.2.1]:1234", "", "192.0.2.1"},
		{"invalid", "", "unknown"},
	} {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.RemoteAddr = test.remote
		if test.forwarded != "" {
			r.Header.Set("X-Forwarded-For", test.forwarded)
		}
		if got := clientAddress(r); got != test.want {
			t.Fatalf("remote=%s forwarded=%q: got %s, want %s", test.remote, test.forwarded, got, test.want)
		}
	}
}

func TestRunnerLimits(t *testing.T) {
	previous := config.App
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
		config.App = previous
	})
	config.App.WASMRuntimeDir = t.TempDir()
	config.App.WASMCacheDir = t.TempDir()
	config.App.WASMLimits.MaxConcurrentTasks = 1
	config.App.WASMLimits.MaxQueuedTasks = 0
	config.App.WASMLimits.MaxTasksPerClient = 1
	config.App.WASMLimits.MaxRequestBytes = 128
	config.App.WASMLimits.MaxSourceBytes = 6
	// An empty function for accepted requests, and a loop for holding one task slot.
	module := []byte{0, 'a', 's', 'm', 1, 0, 0, 0, 1, 4, 1, 0x60, 0, 0, 3, 2, 1, 0, 7, 10, 1, 6, '_', 's', 't', 'a', 'r', 't', 0, 0}
	for name, body := range map[string][]byte{
		"python.wasm":     {10, 4, 1, 2, 0, 0x0b},
		"javascript.wasm": {10, 9, 1, 7, 0, 0x03, 0x40, 0x0c, 0, 0x0b, 0x0b},
	} {
		if err := os.WriteFile(filepath.Join(config.App.WASMRuntimeDir, name), append(append([]byte{}, module...), body...), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	f := flamego.New()
	f.Use(limitRequestBody, func(ctx flamego.Context) {
		ctx.Map(appcontext.Context{
			Context: ctx,
		})
		ctx.Map(&db.Sandbox{
			Template: &db.Tpl{
				Language: []string{"python", "javascript"},
			},
		})
	})
	f.Post("/", NewRunnerHandler().Execute)
	const holder, other = "198.51.100.1:1000", "198.51.100.2:1000"
	request := func(ctx context.Context, remote, language, code string) *httptest.ResponseRecorder {
		body := url.Values{"lang": {language}, "code": {code}}.Encode()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)).WithContext(ctx)
		r.RemoteAddr = remote
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		f.ServeHTTP(w, r)
		return w
	}
	executions := useFakeExecutions(t)
	for _, test := range []struct {
		code   string
		status int
		record db.ExecutionStatus
	}{
		{"123456", http.StatusOK, db.ExecutionStatusSuccess},
		{"中文", http.StatusOK, db.ExecutionStatusSuccess},
		{"1234567", http.StatusRequestEntityTooLarge, db.ExecutionStatusRejected},
		{"中文a", http.StatusRequestEntityTooLarge, db.ExecutionStatusRejected},
	} {
		w := request(context.Background(), other, "python", test.code)
		if w.Code != test.status {
			t.Fatalf("source %q: status=%d body=%s", test.code, w.Code, w.Body.String())
		}
		record := executions.last()
		wantCode := test.code
		if test.record == db.ExecutionStatusRejected {
			wantCode = ""
		}
		if record.Status != test.record || record.Code != wantCode || record.CodeBytes != len(test.code) {
			t.Fatalf("source %q: unexpected record %+v", test.code, record)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ctx.Err() == nil {
			if request(ctx, holder, "javascript", "loop").Code != http.StatusServiceUnavailable {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		w := request(context.Background(), other, "python", "ok")
		if w.Code == http.StatusServiceUnavailable {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected busy response, got %d: %s", w.Code, w.Body.String())
		}
		time.Sleep(time.Millisecond)
	}
	if w := request(context.Background(), holder, "python", "ok"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("client holding a task slot was not limited: %d %s", w.Code, w.Body.String())
	}
	cancel()
	<-done
	if w := request(context.Background(), holder, "python", "ok"); w.Code != http.StatusOK {
		t.Fatalf("cancelled request retained task slot: %d %s", w.Code, w.Body.String())
	}
}
