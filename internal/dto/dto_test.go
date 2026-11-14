// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dbutil"
)

func TestResponses(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	template := &db.Tpl{
		Model: dbutil.Model{
			ID:        7,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  updatedAt,
				Valid: true,
			},
		},
		Name:           "WASM",
		Language:       []string{"python", "go"},
		Timeout:        60,
		MaxMemory:      512,
		InternetAccess: true,
	}
	sandbox := &db.Sandbox{
		Model: dbutil.Model{
			ID:        9,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  updatedAt,
				Valid: true,
			},
		},
		UID:         "sandbox-uid",
		Name:        "Example",
		TemplateID:  7,
		Template:    template,
		Placeholder: "print('hello')",
		Editable:    true,
	}
	wantTemplate := `{"id":7,"createdAt":"2026-01-02T03:04:05.123456Z","updatedAt":"2026-01-02T04:04:05.123456Z","name":"WASM","language":["python","go"],"timeout":60,"maxMemory":512,"internetAccess":true}`
	wantSandbox := `{"id":9,"createdAt":"2026-01-02T03:04:05.123456Z","updatedAt":"2026-01-02T04:04:05.123456Z","uid":"sandbox-uid","name":"Example","templateID":7,"template":` + wantTemplate + `,"placeholder":"print('hello')","editable":true}`

	for _, test := range []struct {
		name  string
		value interface{}
		want  string
	}{
		{"template", ToTemplate(template), wantTemplate},
		{"sandbox with template", ToSandbox(sandbox), wantSandbox},
		{"nil template", ToTemplate(nil), `null`},
		{"nil sandbox", ToSandbox(nil), `null`},
		{"zero template", ToTemplate(&db.Tpl{}), `{"id":0,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","name":"","language":null,"timeout":0,"maxMemory":0,"internetAccess":false}`},
		{"unloaded template", ToSandbox(&db.Sandbox{}), `{"id":0,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","uid":"","name":"","templateID":0,"template":null,"placeholder":"","editable":false}`},
		{"template list", ListTemplatesResp{Templates: []*Template{ToTemplate(template)}, Total: 1}, `{"templates":[` + wantTemplate + `],"total":1}`},
		{"sandbox list", ListSandboxesResp{Sandboxes: []*Sandbox{ToSandbox(sandbox)}, Total: 1}, `{"sandboxes":[` + wantSandbox + `],"total":1}`},
		{"empty template list", ListTemplatesResp{Templates: []*Template{}}, `{"templates":[],"total":0}`},
		{"empty sandbox list", ListSandboxesResp{Sandboxes: []*Sandbox{}}, `{"sandboxes":[],"total":0}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}

			var got, want interface{}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("response contract changed:\ngot:  %s\nwant: %s", data, test.want)
			}
		})
	}
}
