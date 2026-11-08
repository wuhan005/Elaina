// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/pkg/errors"

	"github.com/wuhan005/Elaina/internal/languages"
)

type CreateTemplate struct {
	Name string `json:"name" valid:"required"`
	// Supported languages: c, go, php, python, javascript.
	Language       []string `json:"language" valid:"required"`
	Timeout        int      `json:"timeout,omitempty" valid:"min:0;max:60" minimum:"0" maximum:"60" default:"0"`
	MaxMemory      int64    `json:"maxMemory,omitempty" valid:"min:0;max:4096" minimum:"0" maximum:"4096" default:"0"`
	InternetAccess bool     `json:"internetAccess,omitempty" default:"false"`

	// Decode retired fields to reject unsupported settings from older clients.
	MaxCPUs           int64             `json:"maxCpus" swaggerignore:"true"`
	DNS               map[string]string `json:"dns" swaggerignore:"true"`
	MaxContainer      int64             `json:"maxContainer" swaggerignore:"true"`
	MaxContainerPerIP int64             `json:"maxContainerPerIp" swaggerignore:"true"`
} // @name CreateTemplateReq

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

type UpdateTemplate CreateTemplate // @name UpdateTemplateReq

func (f UpdateTemplate) Validate() error {
	return CreateTemplate(f).Validate()
}
