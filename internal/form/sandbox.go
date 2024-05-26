// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

type CreateSandbox struct {
	Name        string `json:"name" valid:"required"`
	TemplateID  uint   `json:"templateID" valid:"required"`
	Placeholder string `json:"placeholder"`
	Editable    bool   `json:"editable"`
}

type UpdateSandbox struct {
	Name        string `json:"name" valid:"required"`
	TemplateID  uint   `json:"templateID" valid:"required"`
	Placeholder string `json:"placeholder"`
	Editable    bool   `json:"editable"`
}
