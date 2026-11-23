// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"math"
	"net/netip"
	"net/url"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type WASMLimits struct {
	MaxConcurrentTasks      int   `mapstructure:"max_concurrent_tasks"`
	MaxQueuedTasks          int   `mapstructure:"max_queued_tasks"`
	MaxTasksPerClient       int   `mapstructure:"max_tasks_per_client"`
	MaxMemoryMiB            int64 `mapstructure:"max_memory_mib"`
	MaxRequestBytes         int64 `mapstructure:"max_request_bytes"`
	MaxBufferedRequestBytes int64 `mapstructure:"max_buffered_request_bytes"`
	MaxSourceBytes          int64 `mapstructure:"max_source_bytes"`
	MaxWorkspaceBytes       int64 `mapstructure:"max_workspace_bytes"`
	MaxFileBytes            int64 `mapstructure:"max_file_bytes"`
	MaxFiles                int   `mapstructure:"max_files"`
}

var defaultWASMLimits = WASMLimits{
	MaxConcurrentTasks:      1,
	MaxQueuedTasks:          8,
	MaxTasksPerClient:       1,
	MaxMemoryMiB:            512,
	MaxRequestBytes:         4 * 1024 * 1024,
	MaxBufferedRequestBytes: 64 * 1024 * 1024,
	MaxSourceBytes:          1024 * 1024,
	MaxWorkspaceBytes:       64 * 1024 * 1024,
	MaxFileBytes:            32 * 1024 * 1024,
	MaxFiles:                256,
}

const (
	CodeMirrorLayoutCDNJS = "cdnjs"
	CodeMirrorLayoutNPM   = "npm"
)

type CodeMirrorOptions struct {
	BaseURL string `mapstructure:"base_url"`
	Layout  string `mapstructure:"layout"`
}

var App = struct {
	PasswordHash   string            `mapstructure:"password_hash"`
	TrustedProxies []netip.Prefix    `mapstructure:"trusted_proxies"`
	ClientIPHeader string            `mapstructure:"client_ip_header"`
	CodeMirror     CodeMirrorOptions `mapstructure:"codemirror"`

	WASMRuntimeDir           string         `mapstructure:"wasm_runtime_dir"`
	WASMCacheDir             string         `mapstructure:"wasm_cache_dir"`
	WASMKeepCompiledInMemory bool           `mapstructure:"wasm_keep_compiled_in_memory"`
	WASMLimits               WASMLimits     `mapstructure:"wasm_limits"`
	WASMDeniedCIDRs          []netip.Prefix `mapstructure:"wasm_denied_cidrs"`

	// ExecutionRetentionDays is how long execution records are kept. Zero keeps them forever.
	ExecutionRetentionDays int `mapstructure:"execution_retention_days"`
}{
	ClientIPHeader: "X-Forwarded-For",
	CodeMirror: CodeMirrorOptions{
		BaseURL: "https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/",
		Layout:  CodeMirrorLayoutCDNJS,
	},
	WASMRuntimeDir:         ".wasm",
	WASMCacheDir:           ".wasm/cache",
	WASMLimits:             defaultWASMLimits,
	ExecutionRetentionDays: 30,
}

var defaultApp = App

var Postgres struct {
	DSN string `mapstructure:"dsn"`
}

func Init(configFilePath string) error {
	v := viper.New()
	v.SetConfigFile(configFilePath)
	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "read config file")
	}

	// Fields omitted from the config file keep their default values.
	App, Postgres.DSN = defaultApp, ""
	if err := v.UnmarshalKey("app", &App, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(c.DecodeHook, mapstructure.StringToNetIPPrefixHookFunc())
	}); err != nil {
		return errors.Wrap(err, "parse app")
	}

	if err := normalizePrefixes("app.wasm_denied_cidrs", App.WASMDeniedCIDRs); err != nil {
		return err
	}
	if err := normalizePrefixes("app.trusted_proxies", App.TrustedProxies); err != nil {
		return err
	}

	if App.CodeMirror.Layout == CodeMirrorLayoutNPM && !v.IsSet("app.codemirror.base_url") {
		return errors.New("app.codemirror.base_url is required with the npm layout")
	}
	if err := App.CodeMirror.normalize(); err != nil {
		return err
	}

	if err := v.UnmarshalKey("postgres", &Postgres); err != nil {
		return errors.Wrap(err, "parse postgres")
	}

	if _, err := bcrypt.Cost([]byte(App.PasswordHash)); err != nil {
		return errors.New("app.password_hash must contain a bcrypt password hash")
	}
	if App.ExecutionRetentionDays < 0 {
		return errors.New("app.execution_retention_days must not be negative")
	}

	return App.WASMLimits.Validate()
}

func (c *CodeMirrorOptions) normalize() error {
	if c.Layout != CodeMirrorLayoutCDNJS && c.Layout != CodeMirrorLayoutNPM {
		return errors.New("app.codemirror.layout must be cdnjs or npm")
	}

	const invalidBase = "app.codemirror.base_url must be an HTTP(S) URL or a path starting with one slash, without credentials, query or fragment"
	base := strings.TrimSpace(c.BaseURL)
	u, err := url.Parse(base)
	if err != nil || u.User != nil || strings.ContainsAny(base, "?#\\") {
		return errors.New(invalidBase)
	}

	absolute := (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
	local := strings.HasPrefix(base, "/") && !strings.HasPrefix(base, "//")
	if !absolute && !local {
		return errors.New(invalidBase)
	}

	c.BaseURL = strings.TrimRight(u.String(), "/") + "/"
	return nil
}

func normalizePrefixes(key string, prefixes []netip.Prefix) error {
	for i, prefix := range prefixes {
		if !prefix.IsValid() {
			return errors.Errorf("%s[%d]: expected a CIDR", key, i)
		}
		if prefix.Addr().Is4In6() {
			return errors.Errorf("%s[%d]: use IPv4 notation for %q", key, i, prefix)
		}
		prefixes[i] = prefix.Masked()
	}
	return nil
}

func (l WASMLimits) Validate() error {
	if l.MaxConcurrentTasks <= 0 || l.MaxTasksPerClient <= 0 || l.MaxFiles <= 0 {
		return errors.New("app.wasm_limits.max_concurrent_tasks, max_tasks_per_client and max_files must be positive")
	}
	if l.MaxQueuedTasks < 0 {
		return errors.New("app.wasm_limits.max_queued_tasks must not be negative")
	}
	if l.MaxBufferedRequestBytes < l.MaxRequestBytes {
		return errors.New("app.wasm_limits.max_buffered_request_bytes must be at least max_request_bytes")
	}
	if l.MaxMemoryMiB <= 0 || l.MaxMemoryMiB > 4096 {
		return errors.New("app.wasm_limits.max_memory_mib must be between 1 and 4096")
	}
	if l.MaxRequestBytes <= 0 || l.MaxRequestBytes == math.MaxInt64 || l.MaxSourceBytes <= 0 ||
		l.MaxFileBytes <= 0 || l.MaxWorkspaceBytes <= 0 {
		return errors.New("app.wasm_limits byte limits must be positive and max_request_bytes must be below MaxInt64")
	}
	if l.MaxSourceBytes > l.MaxRequestBytes || l.MaxSourceBytes > l.MaxFileBytes || l.MaxFileBytes > l.MaxWorkspaceBytes {
		return errors.New("app.wasm_limits requires max_source_bytes <= max_request_bytes and max_source_bytes <= max_file_bytes <= max_workspace_bytes")
	}
	return nil
}
