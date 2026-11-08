// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var App struct {
	Password       string `envconfig:"APP_PASSWORD"`
	WASMRuntimeDir string `envconfig:"WASM_RUNTIME_DIR" default:".wasm"`
	WASMCacheDir   string `envconfig:"WASM_CACHE_DIR" default:".wasm/cache"`
}

var Postgres struct {
	DSN string `envconfig:"POSTGRES_DSN"`
}

func Init() error {
	if err := envconfig.Process("", &App); err != nil {
		return errors.Wrap(err, "parse app")
	}
	if err := envconfig.Process("", &Postgres); err != nil {
		return errors.Wrap(err, "parse postgres")
	}

	return nil
}
