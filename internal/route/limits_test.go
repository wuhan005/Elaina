// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestRunnerLimits(t *testing.T) {
	previous := config.App
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
		config.App = previous
	})
	config.App.WASMRuntimeDir = t.TempDir()
	config.App.WASMCacheDir = t.TempDir()
	config.App.WASMLimits.MaxConcurrentTasks = 1
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
	request := func(ctx context.Context, language, code string) *httptest.ResponseRecorder {
		body := url.Values{"lang": {language}, "code": {code}}.Encode()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)).WithContext(ctx)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		f.ServeHTTP(w, r)
		return w
	}
	for _, test := range []struct {
		code   string
		status int
	}{
		{"123456", http.StatusOK},
		{"中文", http.StatusOK},
		{"1234567", http.StatusRequestEntityTooLarge},
		{"中文a", http.StatusRequestEntityTooLarge},
	} {
		w := request(context.Background(), "python", test.code)
		if w.Code != test.status {
			t.Fatalf("source %q: status=%d body=%s", test.code, w.Code, w.Body.String())
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ctx.Err() == nil {
			if request(ctx, "javascript", "loop").Code != http.StatusServiceUnavailable {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		w := request(context.Background(), "python", "ok")
		if w.Code == http.StatusServiceUnavailable {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected busy response, got %d: %s", w.Code, w.Body.String())
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	if w := request(context.Background(), "python", "ok"); w.Code != http.StatusOK {
		t.Fatalf("cancelled request retained task slot: %d %s", w.Code, w.Body.String())
	}
}
