// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/config"
)

func newTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN to run database integration tests")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	adminConnection, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adminConnection.Close() })

	namespace := fmt.Sprintf("elaina_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + namespace).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP SCHEMA " + namespace + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	})

	previousConfig, templates, sandboxes := config.Postgres, Tpls, Sandboxes
	t.Cleanup(func() { config.Postgres, Tpls, Sandboxes = previousConfig, templates, sandboxes })
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		address, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		query := address.Query()
		query.Set("search_path", namespace)
		address.RawQuery = query.Encode()
		config.Postgres.DSN = address.String()
	} else {
		config.Postgres.DSN = dsn + " search_path=" + namespace
	}

	database, err := Init()
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	return database
}

func TestInit(t *testing.T) {
	newTestDatabase(t)
	ctx := context.Background()
	template, err := Tpls.Create(ctx, CreateTplOptions{
		Name:     "Timestamp precision",
		Language: []string{"python"},
	})
	if err != nil {
		t.Fatal(err)
	}

	stored, err := Tpls.GetByID(ctx, template.ID)
	if err != nil {
		t.Fatal(err)
	}
	if template.CreatedAt.IsZero() || !stored.CreatedAt.Equal(template.CreatedAt) || !stored.UpdatedAt.Equal(template.UpdatedAt) {
		t.Fatal("database round trip changed the generated timestamps")
	}

	sandbox, err := Sandboxes.Create(ctx, CreateSandboxOptions{
		Name:       "Migrated sandbox table",
		TemplateID: template.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	storedSandbox, err := Sandboxes.GetByID(ctx, sandbox.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedSandbox.Template == nil || storedSandbox.Template.ID != template.ID {
		t.Fatal("sandbox template association was not loaded")
	}
}

func TestInitDropsRetiredTemplateColumns(t *testing.T) {
	database := newTestDatabase(t)
	ctx := context.Background()
	template, err := Tpls.Create(ctx, CreateTplOptions{
		Name:           "Existing template",
		Language:       []string{"go"},
		Timeout:        60,
		MaxMemory:      512,
		InternetAccess: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	sandbox, err := Sandboxes.Create(ctx, CreateSandboxOptions{Name: "Existing sandbox", TemplateID: template.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`ALTER TABLE tpls
		ADD COLUMN max_cpu bigint DEFAULT 2,
		ADD COLUMN max_cpus bigint DEFAULT 4,
		ADD COLUMN max_container bigint DEFAULT 5,
		ADD COLUMN max_container_per_ip bigint DEFAULT 3,
		ADD COLUMN dns jsonb DEFAULT '{"example.com":"127.0.0.1"}'`).Error; err != nil {
		t.Fatal(err)
	}

	for range 2 {
		upgraded, err := Init()
		if err != nil {
			t.Fatal(err)
		}
		connection, err := upgraded.DB()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = connection.Close() })
		for _, column := range []string{"max_cpu", "max_cpus", "max_container", "max_container_per_ip", "dns"} {
			if upgraded.Migrator().HasColumn(&Tpl{}, column) {
				t.Errorf("retired column %s still exists", column)
			}
		}
	}

	storedSandbox, err := Sandboxes.GetByID(ctx, sandbox.ID)
	if err != nil {
		t.Fatal(err)
	}
	stored := storedSandbox.Template
	if stored == nil || stored.ID != template.ID || stored.Name != template.Name || len(stored.Language) != 1 ||
		stored.Language[0] != "go" || stored.Timeout != 60 || stored.MaxMemory != 512 || !stored.InternetAccess {
		t.Fatalf("migration changed the sandbox's template: %+v", stored)
	}
}
