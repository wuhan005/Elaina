// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/pkg/errors"

	"github.com/wuhan005/Elaina/internal/config"
)

type CreateSandbox struct {
	Name        string `json:"name" valid:"required"`
	TemplateID  uint   `json:"templateID" valid:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	Editable    bool   `json:"editable,omitempty"`
} // @name CreateSandboxReq

type UpdateSandbox struct {
	Name        string `json:"name" valid:"required"`
	TemplateID  uint   `json:"templateID" valid:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	Editable    bool   `json:"editable,omitempty"`
} // @name UpdateSandboxReq

func (f CreateSandbox) Validate() error {
	if int64(len(f.Placeholder)) > config.App.WASMLimits.MaxSourceBytes {
		return errors.New("placeholder exceeds the configured source size limit")
	}
	return nil
}

func (f UpdateSandbox) Validate() error {
	return CreateSandbox(f).Validate()
}
