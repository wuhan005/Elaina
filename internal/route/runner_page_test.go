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

	"github.com/wuhan005/Elaina/internal/config"
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
	previous, defaultCodeMirror := db.Sandboxes, config.App.CodeMirror
	db.Sandboxes = runnerPageStore{sandbox: &db.Sandbox{
		UID:         "example",
		Name:        `Example "sandbox"`,
		Placeholder: placeholder,
		Template:    &db.Tpl{Language: []string{"python", "javascript"}},
	}}
	t.Cleanup(func() { db.Sandboxes, config.App.CodeMirror = previous, defaultCodeMirror })

	for _, tc := range []struct {
		name, method, code, language string
		source                       config.CodeMirrorOptions
		base, core, suffix           string
	}{
		{
			name: "cdnjs", method: http.MethodGet, code: placeholder, language: "python",
			source: defaultCodeMirror,
			base:   "https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/",
			core:   "codemirror", suffix: ".min",
		},
		{
			name: "npm", method: http.MethodPost, language: "javascript",
			code:   "function f() {\n\treturn '  </textarea> & {{ .Code }}  ';\n}\n",
			source: config.CodeMirrorOptions{BaseURL: "https://cdn.jsdelivr.net/npm/codemirror@5.65.16/", Layout: config.CodeMirrorLayoutNPM},
			base:   "https://cdn.jsdelivr.net/npm/codemirror@5.65.16/",
			core:   "lib/codemirror",
		},
		{
			name: "same origin", method: http.MethodGet, code: placeholder, language: "python",
			source: config.CodeMirrorOptions{BaseURL: "/vendor/codemirror/5.65.16/", Layout: config.CodeMirrorLayoutCDNJS},
			base:   "/vendor/codemirror/5.65.16/",
			core:   "codemirror", suffix: ".min",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.App.CodeMirror = tc.source
			router := newTestClient(t).router

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

			if css := tc.base + tc.core + tc.suffix + ".css"; !strings.Contains(body, `href="`+css+`"`) {
				t.Fatalf("missing stylesheet %q", css)
			}

			paths := []string{tc.core, "mode/xml/xml", "mode/htmlmixed/htmlmixed", "mode/clike/clike", "mode/php/php", "mode/python/python", "mode/go/go", "mode/javascript/javascript"}
			if count := strings.Count(body, `<script src="`); count != len(paths)+1 {
				t.Fatalf("expected %d external scripts, got %d", len(paths)+1, count)
			}

			remaining := body
			for _, path := range paths {
				script := tc.base + path + tc.suffix + ".js"
				_, tail, found := strings.Cut(remaining, `src="`+script+`"`)
				if !found {
					t.Fatalf("missing or misplaced script %q", script)
				}
				remaining = tail
			}
			if !strings.Contains(remaining, `src="/static/js/sandbox.js"`) {
				t.Fatal("sandbox.js must load after CodeMirror and its modes")
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
