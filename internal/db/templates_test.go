// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
)

func TestTemplateSettings(t *testing.T) {
	newTestDatabase(t)
	ctx := context.Background()
	tpl, err := Tpls.Create(ctx, CreateTplOptions{
		Name:           "WASM",
		Language:       []string{"python"},
		Timeout:        60,
		MaxMemory:      512,
		InternetAccess: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := Tpls.Update(ctx, tpl.ID, UpdateTplOptions{
		Name:           "Updated",
		Language:       []string{"c", "go"},
		Timeout:        0,
		MaxMemory:      0,
		InternetAccess: false,
	}); err != nil {
		t.Fatal(err)
	}
	tpl, err = Tpls.GetByID(ctx, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "Updated" || len(tpl.Language) != 2 || tpl.Timeout != 0 || tpl.MaxMemory != 0 || tpl.InternetAccess {
		t.Fatalf("zero-valued settings were not saved: %+v", tpl)
	}
}
