// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/pkg/errors"

	"github.com/wuhan005/Elaina/internal/languages"
)

type CreateTemplate struct {
	Name           string   `json:"name" valid:"required"`
	Language       []string `json:"language" valid:"required"`
	Timeout        int      `json:"timeout" valid:"min:0;max:60"`
	MaxMemory      int64    `json:"maxMemory" valid:"min:0;max:4096"`
	InternetAccess bool     `json:"internetAccess"`

	// Decode retired fields to reject unsupported settings from older clients.
	MaxCPUs           int64             `json:"maxCpus"`
	DNS               map[string]string `json:"dns"`
	MaxContainer      int64             `json:"maxContainer"`
	MaxContainerPerIP int64             `json:"maxContainerPerIp"`
}

func (f CreateTemplate) Validate() error {
	if f.MaxCPUs != 0 {
		return errors.New("WASM does not support per-task CPU quotas; omit maxCpus or set it to 0")
	}
	if f.MaxContainer != 0 || f.MaxContainerPerIP != 0 {
		return errors.New("container count limits are no longer supported; concurrency limits are not enforced")
	}
	if len(f.DNS) != 0 {
		return errors.New("custom DNS settings are not supported; omit dns or use an empty object")
	}
	for _, name := range f.Language {
		if _, ok := languages.Get(name); !ok {
			return errors.Errorf("unsupported WASM language: %q", name)
		}
	}
	return nil
}

type UpdateTemplate CreateTemplate

func (f UpdateTemplate) Validate() error {
	return CreateTemplate(f).Validate()
}
