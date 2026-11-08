// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var App struct {
	Password       string `mapstructure:"password"`
	WASMRuntimeDir string `mapstructure:"wasm_runtime_dir"`
	WASMCacheDir   string `mapstructure:"wasm_cache_dir"`
}

var Postgres struct {
	DSN string `mapstructure:"dsn"`
}

func Init(configFilePath string) error {
	v := viper.New()
	v.SetConfigFile(configFilePath)
	v.SetDefault("app.wasm_runtime_dir", ".wasm")
	v.SetDefault("app.wasm_cache_dir", ".wasm/cache")
	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "read config file")
	}

	if err := v.UnmarshalKey("app", &App); err != nil {
		return errors.Wrap(err, "parse app")
	}
	App.WASMRuntimeDir = v.GetString("app.wasm_runtime_dir")
	App.WASMCacheDir = v.GetString("app.wasm_cache_dir")

	if err := v.UnmarshalKey("postgres", &Postgres); err != nil {
		return errors.Wrap(err, "parse postgres")
	}

	return nil
}
