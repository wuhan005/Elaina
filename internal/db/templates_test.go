// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
)

func TestTemplateSettings(t *testing.T) {
	database := newTestDatabase(t)

	// Simulate columns left behind by an existing container-based deployment.
	table := "tpls"
	if err := database.Exec("ALTER TABLE " + table + " ADD COLUMN max_container bigint, ADD COLUMN max_container_per_ip bigint, ADD COLUMN dns jsonb").Error; err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&Tpl{}); err != nil {
		t.Fatal(err)
	}
	store := NewTplStore(database)
	ctx := context.Background()
	tpl, err := store.Create(ctx, CreateTplOptions{
		Name:           "WASM",
		Language:       []string{"python"},
		Timeout:        60,
		MaxMemory:      512,
		InternetAccess: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Model(tpl).Update("MaxCPU", 2).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Exec("UPDATE "+table+" SET max_container = 5, max_container_per_ip = 3, dns = '{\"example.com\":\"127.0.0.1\"}' WHERE id = ?", tpl.ID).Error; err != nil {
		t.Fatal(err)
	}
	tpl, err = store.GetByID(ctx, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tpl.MaxCPU != 2 {
		t.Fatal("legacy CPU quota was silently discarded")
	}
	if err := store.Update(ctx, tpl.ID, UpdateTplOptions{
		Name:           "Updated",
		Language:       []string{"c", "go"},
		Timeout:        0,
		MaxCPU:         0,
		MaxMemory:      0,
		InternetAccess: false,
	}); err != nil {
		t.Fatal(err)
	}
	tpl, err = store.GetByID(ctx, tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "Updated" || len(tpl.Language) != 2 || tpl.MaxCPU != 0 || tpl.Timeout != 0 || tpl.MaxMemory != 0 || tpl.InternetAccess {
		t.Fatalf("zero-valued settings were not saved: %+v", tpl)
	}
	var legacy struct {
		MaxContainer      int64
		MaxContainerPerIP int64
		DNS               string
	}
	if err := database.Table(table).Select("max_container, max_container_per_ip, dns::text").Where("id = ?", tpl.ID).Scan(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if legacy.MaxContainer != 5 || legacy.MaxContainerPerIP != 3 || legacy.DNS != `{"example.com": "127.0.0.1"}` {
		t.Fatalf("retired database values were changed: %+v", legacy)
	}
}
