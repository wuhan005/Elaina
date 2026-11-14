// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flamego/flamego"

	appcontext "github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/runtime"
)

func TestTemplateBinding(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, test := range []struct {
			name   string
			fields map[string]interface{}
			error  string
		}{
			{"defaults", nil, ""},
			{"upper limits", map[string]interface{}{"timeout": 60, "maxMemory": 4096, "internetAccess": true}, ""},
			{"minimum memory", map[string]interface{}{"maxMemory": 1}, ""},
			{"negative timeout", map[string]interface{}{"timeout": -1}, "Timeout"},
			{"excessive timeout", map[string]interface{}{"timeout": 61}, "Timeout"},
			{"fractional timeout", map[string]interface{}{"timeout": 0.5}, "parse request"},
			{"negative memory", map[string]interface{}{"maxMemory": -1}, "MaxMemory"},
			{"excessive memory", map[string]interface{}{"maxMemory": 4097}, "MaxMemory"},
			{"fractional memory", map[string]interface{}{"maxMemory": 0.5}, "parse request"},
			{"missing name", map[string]interface{}{"name": ""}, "Name"},
			{"empty languages", map[string]interface{}{"language": []string{}}, "Language"},
			{"unsupported language", map[string]interface{}{"language": []string{"rust"}}, "unsupported WASM language"},
		} {
			t.Run(method+"/"+test.name, func(t *testing.T) {
				body := map[string]interface{}{"name": "WASM", "language": []string{"c", "go", "python", "php", "javascript"}}
				for key, value := range test.fields {
					body[key] = value
				}
				data, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				f := flamego.New()
				f.Use(func(c flamego.Context) {
					c.Map(appcontext.Context{
						Context: c,
					})
				})
				called := false
				accept := func(c appcontext.Context, value CreateTemplate) {
					called = true
					for _, language := range value.Language {
						_, err := runtime.NewTask(context.Background(), runtime.NewTaskOptions{
							Language: language,
							Template: &db.Tpl{
								Timeout:        value.Timeout,
								MaxMemory:      value.MaxMemory,
								InternetAccess: value.InternetAccess,
							},
						})
						if err != nil {
							t.Errorf("accepted settings rejected by runtime: %v", err)
						}
					}
					c.Status(http.StatusNoContent)
				}
				f.Post("/", Bind(CreateTemplate{}), accept)
				f.Put("/", Bind(UpdateTemplate{}), func(c appcontext.Context, value UpdateTemplate) { accept(c, CreateTemplate(value)) })
				response := httptest.NewRecorder()
				f.ServeHTTP(response, httptest.NewRequest(method, "/", bytes.NewReader(data)))
				if test.error == "" {
					if !called || response.Code != http.StatusNoContent {
						t.Fatalf("valid template: status=%d body=%s", response.Code, response.Body)
					}
				} else if called || response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), test.error) {
					t.Fatalf("invalid template reached handler or wrong error: status=%d body=%s", response.Code, response.Body)
				}
			})
		}
	}
}
