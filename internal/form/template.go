// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/wuhan005/Elaina/internal/i18n"
	"github.com/wuhan005/Elaina/internal/languages"
)

type CreateTemplate struct {
	Name string `json:"name" valid:"required"`
	// Supported languages: c, go, php, python, javascript.
	Language       []string `json:"language" valid:"required"`
	Timeout        int      `json:"timeout,omitempty" valid:"min:0;max:60" minimum:"0" maximum:"60" default:"0"`
	MaxMemory      int64    `json:"maxMemory,omitempty" valid:"min:0;max:4096" minimum:"0" maximum:"4096" default:"0"`
	InternetAccess bool     `json:"internetAccess,omitempty" default:"false"`
} // @name CreateTemplateReq

func (f CreateTemplate) Validate() error {
	for _, name := range f.Language {
		if _, ok := languages.Get(name); !ok {
			return i18n.Errorf("validation::unsupported_language", name)
		}
	}
	return nil
}

type UpdateTemplate CreateTemplate // @name UpdateTemplateReq

func (f UpdateTemplate) Validate() error {
	return CreateTemplate(f).Validate()
}
