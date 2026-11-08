// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	t.Setenv("APP_PASSWORD", "environment-password")
	t.Setenv("POSTGRES_DSN", "environment-dsn")
	t.Setenv("WASM_RUNTIME_DIR", "environment-runtime")
	t.Setenv("WASM_CACHE_DIR", "environment-cache")

	for _, test := range []struct {
		name, content, password, runtimeDir, cacheDir, dsn string
	}{
		{
			"file values", `app:
  password: file-password
  wasm_runtime_dir: /opt/elaina/wasm
  wasm_cache_dir: /var/cache/elaina
postgres:
  dsn: "postgres://user:password@localhost/elaina?sslmode=disable"
`, "file-password", "/opt/elaina/wasm", "/var/cache/elaina", "postgres://user:password@localhost/elaina?sslmode=disable",
		},
		{
			"default directories", "app:\n  password: another-password\npostgres:\n  dsn: file-dsn\n",
			"another-password", ".wasm", ".wasm/cache", "file-dsn",
		},
		{"partial sections reset credentials", "app:\n  wasm_runtime_dir: ./local-wasm\npostgres: {}\n", "", "./local-wasm", ".wasm/cache", ""},
		{"omitted sections reset previous values", "{}\n", "", ".wasm", ".wasm/cache", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := Init(writeConfig(t, test.content)); err != nil {
				t.Fatal(err)
			}
			if App.Password != test.password || App.WASMRuntimeDir != test.runtimeDir || App.WASMCacheDir != test.cacheDir || Postgres.DSN != test.dsn {
				t.Fatal("configuration does not match the file and defaults")
			}
		})
	}
}

func TestInitErrors(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	if err := Init(writeConfig(t, "app:\n  password: original-password\npostgres:\n  dsn: original-dsn\n")); err != nil {
		t.Fatal(err)
	}
	previousApp, previousPostgres := App, Postgres
	for _, test := range []struct{ name, content, error string }{
		{"missing file", "", "read config file"},
		{"invalid YAML", "app: [\n", "read config file"},
		{"invalid app section", "app: invalid\n", "parse app"},
		{"invalid postgres section", "app:\n  password: changed-password\npostgres: invalid\n", "parse postgres"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "missing.yaml")
			if test.content != "" {
				path = writeConfig(t, test.content)
			}
			err := Init(path)
			if err == nil || !strings.Contains(err.Error(), test.error) {
				t.Fatalf("expected %q, got %v", test.error, err)
			}
			if App != previousApp || Postgres != previousPostgres {
				t.Fatal("failed initialization changed the active configuration")
			}
		})
	}
}

func TestExampleConfig(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	if err := Init("../../config/elaina.example.yaml"); err != nil {
		t.Fatal(err)
	}
	if App.Password != "" || App.WASMRuntimeDir != ".wasm" || App.WASMCacheDir != ".wasm/cache" || Postgres.DSN == "" {
		t.Fatal("unexpected example configuration")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "elaina.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
