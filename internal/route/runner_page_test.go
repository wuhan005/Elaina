package route

import (
	"bytes"
	"context"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/public"
)

type runnerPageStore struct {
	db.SandboxStore
	sandbox *db.Sandbox
}

func (s runnerPageStore) GetByUID(_ context.Context, uid string) (*db.Sandbox, error) {
	if uid != s.sandbox.UID {
		return nil, db.ErrSandboxNotFound
	}
	return s.sandbox, nil
}

func TestMinifiedRunnerPage(t *testing.T) {
	const placeholder = "if True:\n    print('中文')\n"
	previous := db.Sandboxes
	db.Sandboxes = runnerPageStore{sandbox: &db.Sandbox{
		UID:         "example",
		Name:        `Example "sandbox"`,
		Placeholder: placeholder,
		Template:    &db.Tpl{Language: []string{"python", "javascript"}},
	}}
	t.Cleanup(func() { db.Sandboxes = previous })

	router := newTestClient(t).router

	for _, tc := range []struct {
		method, code, language string
	}{
		{http.MethodGet, placeholder, "python"},
		{http.MethodPost, "function f() {\n\treturn '  </textarea> & {{ .Code }}  ';\n}\n", "javascript"},
	} {
		t.Run(tc.method, func(t *testing.T) {
			var form string
			if tc.method == http.MethodPost {
				form = url.Values{"c": {tc.code}}.Encode()
			}

			req := httptest.NewRequest(tc.method, "/r/example?l="+tc.language+"&lang=zh-CN", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			if res.Code != http.StatusOK || res.Header().Get("Content-Type") != "text/html; charset=utf-8" {
				t.Fatalf("status=%d headers=%v body=%s", res.Code, res.Header(), res.Body)
			}

			body := res.Body.String()
			_, code, hasStart := strings.Cut(body, `<textarea id="code">`)
			code, _, hasEnd := strings.Cut(code, `</textarea>`)
			if !hasStart || !hasEnd || html.UnescapeString(code) != tc.code {
				t.Fatalf("unexpected editor code %q", code)
			}

			if !strings.Contains(body, `<html lang="zh-CN">`) || !strings.Contains(body, `"language":"`+tc.language+`"`) {
				t.Fatalf("missing page language or runner configuration: %s", body)
			}

			if strings.Contains(body, "// Pick the theme") || strings.Contains(body, "\n  <meta") {
				t.Fatal("page template was not minified")
			}
		})
	}
}

func TestMinifiedRunnerScript(t *testing.T) {
	router := newTestClient(t).router
	const path = "/static/js/sandbox.js"
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
	if get.Code != http.StatusOK || !strings.Contains(get.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("GET status=%d headers=%v", get.Code, get.Header())
	}

	original, err := public.FS.ReadFile(sandboxScriptPath)
	if err != nil {
		t.Fatal(err)
	}
	if get.Body.Len() < 12 || get.Body.Len() >= len(original) {
		t.Fatalf("unexpected script size %d, original %d", get.Body.Len(), len(original))
	}
	t.Logf("script size: %d -> %d bytes", len(original), get.Body.Len())

	head := httptest.NewRecorder()
	router.ServeHTTP(head, httptest.NewRequest(http.MethodHead, path, nil))
	if head.Code != http.StatusOK || head.Body.Len() != 0 || head.Header().Get("Content-Length") != strconv.Itoa(get.Body.Len()) {
		t.Fatalf("HEAD status=%d headers=%v bytes=%d", head.Code, head.Header(), head.Body.Len())
	}

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Range", "bytes=0-11")
	partial := httptest.NewRecorder()
	router.ServeHTTP(partial, req)
	if partial.Code != http.StatusPartialContent || !bytes.Equal(partial.Body.Bytes(), get.Body.Bytes()[:12]) {
		t.Fatalf("Range status=%d body=%q", partial.Code, partial.Body.Bytes())
	}
}
