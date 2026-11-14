// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPasswordHash = "$2y$04$d09ks9hqtbe2WG0nD4h51u0w0bPHxtVkUbr0qxmVFrXSY/0P6K2Ay"

func TestInit(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	t.Setenv("APP_PASSWORD", "environment-password")
	t.Setenv("APP_PASSWORD_HASH", "environment-hash")
	t.Setenv("POSTGRES_DSN", "environment-dsn")
	t.Setenv("WASM_RUNTIME_DIR", "environment-runtime")
	t.Setenv("WASM_CACHE_DIR", "environment-cache")

	for _, test := range []struct {
		name, content, runtimeDir, cacheDir, dsn string
	}{
		{
			"file values",
			`app:
  password_hash: "` + testPasswordHash + `"
  wasm_runtime_dir: /opt/elaina/wasm
  wasm_cache_dir: /var/cache/elaina
postgres:
  dsn: "postgres://user:password@localhost/elaina?sslmode=disable"
`,
			"/opt/elaina/wasm",
			"/var/cache/elaina",
			"postgres://user:password@localhost/elaina?sslmode=disable",
		},
		{"default directories", "app:\n  password_hash: " + testPasswordHash + "\npostgres:\n  dsn: file-dsn\n", ".wasm", ".wasm/cache", "file-dsn"},
		{"app only", "app:\n  password_hash: " + testPasswordHash + "\n  wasm_runtime_dir: ./local-wasm\n", "./local-wasm", ".wasm/cache", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			App.PasswordHash, Postgres.DSN = "", ""
			if err := Init(writeConfig(t, test.content)); err != nil {
				t.Fatal(err)
			}
			if App.PasswordHash != testPasswordHash || App.WASMRuntimeDir != test.runtimeDir || App.WASMCacheDir != test.cacheDir || Postgres.DSN != test.dsn {
				t.Fatal("configuration does not match the file and defaults")
			}
		})
	}
}

func TestWASMLimits(t *testing.T) {
	previous := App
	t.Cleanup(func() { App = previous })
	base := "app:\n  password_hash: " + testPasswordHash + "\n"
	for _, test := range []struct{ name, settings string }{
		{"zero concurrency", "max_concurrent_tasks: 0"},
		{"negative concurrency", "max_concurrent_tasks: -1"},
		{"zero memory", "max_memory_mib: 0"},
		{"excessive memory", "max_memory_mib: 4097"},
		{"zero files", "max_files: 0"},
		{"zero body", "max_request_bytes: 0"},
		{"overflow body", "max_request_bytes: 9223372036854775807"},
		{"zero source", "max_source_bytes: 0"},
		{"negative workspace", "max_workspace_bytes: -1"},
		{"zero file", "max_file_bytes: 0"},
		{"source exceeds body", "max_request_bytes: 1"},
		{"source exceeds file", "max_file_bytes: 1"},
		{"file exceeds workspace", "max_workspace_bytes: 1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := Init(writeConfig(t, base+"  wasm_limits:\n    "+test.settings+"\n"))
			if err == nil || !strings.Contains(err.Error(), "app.wasm_limits") {
				t.Fatalf("expected limits validation error, got %v", err)
			}
		})
	}

	custom := WASMLimits{
		MaxConcurrentTasks: 3,
		MaxMemoryMiB:       128,
		MaxRequestBytes:    4096,
		MaxSourceBytes:     1024,
		MaxWorkspaceBytes:  8192,
		MaxFileBytes:       2048,
		MaxFiles:           16,
	}
	settings := fmt.Sprintf("  wasm_limits:\n    max_concurrent_tasks: %d\n    max_memory_mib: %d\n    max_request_bytes: %d\n    max_source_bytes: %d\n    max_workspace_bytes: %d\n    max_file_bytes: %d\n    max_files: %d\n",
		custom.MaxConcurrentTasks, custom.MaxMemoryMiB, custom.MaxRequestBytes, custom.MaxSourceBytes, custom.MaxWorkspaceBytes, custom.MaxFileBytes, custom.MaxFiles)
	if err := Init(writeConfig(t, base+settings)); err != nil || App.WASMLimits != custom {
		t.Fatalf("custom limits: %+v, %v", App.WASMLimits, err)
	}
	if err := Init(writeConfig(t, base)); err != nil || App.WASMLimits != defaultWASMLimits {
		t.Fatalf("default limits: %+v, %v", App.WASMLimits, err)
	}
}

func TestWASMDeniedCIDRs(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	base := "app:\n  password_hash: " + testPasswordHash + "\n"
	for _, value := range []string{"[8.8.8.7/24, '2606:4700::1/32']", "'8.8.8.7/24,2606:4700::1/32'"} {
		if err := Init(writeConfig(t, base+"  wasm_denied_cidrs: "+value+"\n")); err != nil {
			t.Fatal(err)
		}
		if got := App.WASMDeniedCIDRs; len(got) != 2 || got[0] != netip.MustParsePrefix("8.8.8.0/24") || got[1] != netip.MustParsePrefix("2606:4700::/32") {
			t.Fatalf("unexpected denied CIDRs: %v", got)
		}
	}
	if err := Init(writeConfig(t, base)); err != nil || len(App.WASMDeniedCIDRs) != 0 {
		t.Fatalf("expected no additional CIDRs: %v, %v", App.WASMDeniedCIDRs, err)
	}
	for _, value := range []string{"[bad-cidr]", "['::ffff:8.8.8.0/120']", "[null]"} {
		err := Init(writeConfig(t, base+"  wasm_denied_cidrs: "+value+"\n"))
		if err == nil || !strings.Contains(err.Error(), "wasm_denied_cidrs[0]") {
			t.Fatalf("expected error for %s, got %v", value, err)
		}
	}
}

func TestInitErrors(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	for _, test := range []struct{ name, content, error string }{
		{"missing file", "", "read config file"},
		{"invalid YAML", "app: [\n", "read config file"},
		{"invalid app section", "app: invalid\n", "parse app"},
		{"invalid postgres section", "app:\n  password_hash: changed-hash\npostgres: invalid\n", "parse postgres"},
		{"omitted sections", "{}\n", "app.password_hash"},
		{"missing hash", "app:\n  wasm_runtime_dir: ./local-wasm\n", "app.password_hash"},
		{"empty hash", "app:\n  password_hash: \"\"\n", "app.password_hash"},
		{"legacy password", "app:\n  password: plaintext-password\n", "app.password_hash"},
		{"plaintext hash", "app:\n  password_hash: plaintext-password\n", "app.password_hash"},
		{"truncated hash", "app:\n  password_hash: " + testPasswordHash[:30] + "\n", "app.password_hash"},
		{"invalid cost", "app:\n  password_hash: " + strings.Replace(testPasswordHash, "$04$", "$99$", 1) + "\n", "app.password_hash"},
	} {
		t.Run(test.name, func(t *testing.T) {
			App.PasswordHash, Postgres.DSN = "", ""
			path := filepath.Join(t.TempDir(), "missing.yaml")
			if test.content != "" {
				path = writeConfig(t, test.content)
			}
			err := Init(path)
			if err == nil || !strings.Contains(err.Error(), test.error) {
				t.Fatalf("expected %q, got %v", test.error, err)
			}
		})
	}
}

func TestExampleConfig(t *testing.T) {
	app, postgres := App, Postgres
	t.Cleanup(func() { App, Postgres = app, postgres })
	if err := Init("../../config/elaina.example.yaml"); err == nil || !strings.Contains(err.Error(), "app.password_hash") {
		t.Fatalf("expected missing password hash error, got %v", err)
	}
	if App.PasswordHash != "" || App.WASMRuntimeDir != ".wasm" || App.WASMCacheDir != ".wasm/cache" || Postgres.DSN == "" {
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
