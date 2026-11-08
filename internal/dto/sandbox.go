// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

import (
	"time"

	"github.com/wuhan005/Elaina/internal/db"
)

type Sandbox struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	UID         string    `json:"uid"`
	Name        string    `json:"name"`
	TemplateID  uint      `json:"templateID"`
	Template    *Template `json:"template" extensions:"x-nullable"`
	Placeholder string    `json:"placeholder"`
	Editable    bool      `json:"editable"`
} // @name Sandbox

func ToSandbox(sandbox *db.Sandbox) *Sandbox {
	if sandbox == nil {
		return nil
	}

	return &Sandbox{
		ID:          sandbox.ID,
		CreatedAt:   sandbox.CreatedAt,
		UpdatedAt:   sandbox.UpdatedAt,
		UID:         sandbox.UID,
		Name:        sandbox.Name,
		TemplateID:  sandbox.TemplateID,
		Template:    ToTemplate(sandbox.Template),
		Placeholder: sandbox.Placeholder,
		Editable:    sandbox.Editable,
	}
}

type ListSandboxesResp struct {
	Sandboxes []*Sandbox `json:"sandboxes"`
	Total     int64      `json:"total"`
} // @name ListSandboxesResp
