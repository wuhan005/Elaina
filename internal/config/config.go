// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"math"
	"net/netip"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type WASMLimits struct {
	MaxConcurrentTasks int   `mapstructure:"max_concurrent_tasks"`
	MaxMemoryMiB       int64 `mapstructure:"max_memory_mib"`
	MaxRequestBytes    int64 `mapstructure:"max_request_bytes"`
	MaxSourceBytes     int64 `mapstructure:"max_source_bytes"`
	MaxWorkspaceBytes  int64 `mapstructure:"max_workspace_bytes"`
	MaxFileBytes       int64 `mapstructure:"max_file_bytes"`
	MaxFiles           int   `mapstructure:"max_files"`
}

var defaultWASMLimits = WASMLimits{
	MaxConcurrentTasks: 1,
	MaxMemoryMiB:       512,
	MaxRequestBytes:    4 * 1024 * 1024,
	MaxSourceBytes:     1024 * 1024,
	MaxWorkspaceBytes:  64 * 1024 * 1024,
	MaxFileBytes:       32 * 1024 * 1024,
	MaxFiles:           256,
}

var App = struct {
	PasswordHash string `mapstructure:"password_hash"`

	WASMRuntimeDir  string         `mapstructure:"wasm_runtime_dir"`
	WASMCacheDir    string         `mapstructure:"wasm_cache_dir"`
	WASMLimits      WASMLimits     `mapstructure:"wasm_limits"`
	WASMDeniedCIDRs []netip.Prefix `mapstructure:"wasm_denied_cidrs"`
}{
	WASMRuntimeDir: ".wasm",
	WASMCacheDir:   ".wasm/cache",
	WASMLimits:     defaultWASMLimits,
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

	for i, prefix := range App.WASMDeniedCIDRs {
		if !prefix.IsValid() {
			return errors.Errorf("app.wasm_denied_cidrs[%d]: expected a CIDR", i)
		}
		if prefix.Addr().Is4In6() {
			return errors.Errorf("app.wasm_denied_cidrs[%d]: use IPv4 notation for %q", i, prefix)
		}
		App.WASMDeniedCIDRs[i] = prefix.Masked()
	}

	if err := v.UnmarshalKey("postgres", &Postgres); err != nil {
		return errors.Wrap(err, "parse postgres")
	}

	if _, err := bcrypt.Cost([]byte(App.PasswordHash)); err != nil {
		return errors.New("app.password_hash must contain a bcrypt password hash")
	}

	return App.WASMLimits.Validate()
}

func (l WASMLimits) Validate() error {
	if l.MaxConcurrentTasks <= 0 || l.MaxFiles <= 0 {
		return errors.New("app.wasm_limits.max_concurrent_tasks and max_files must be positive")
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
