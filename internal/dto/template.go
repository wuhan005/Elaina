// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

import (
	"time"

	"github.com/wuhan005/Elaina/internal/db"
)

type Template struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name           string   `json:"name"`
	Language       []string `json:"language"`
	Timeout        int      `json:"timeout"`
	MaxMemory      int64    `json:"maxMemory"`
	InternetAccess bool     `json:"internetAccess"`

	// Keep the legacy CPU quota visible until the template is migrated.
	MaxCPUs int64 `json:"maxCpus"`
}

func ToTemplate(template *db.Tpl) *Template {
	if template == nil {
		return nil
	}

	return &Template{
		ID:             template.ID,
		CreatedAt:      template.CreatedAt,
		UpdatedAt:      template.UpdatedAt,
		Name:           template.Name,
		Language:       template.Language,
		Timeout:        template.Timeout,
		MaxMemory:      template.MaxMemory,
		InternetAccess: template.InternetAccess,
		MaxCPUs:        template.MaxCPUs,
	}
}

type ListTemplatesResp struct {
	Templates []*Template `json:"templates"`
	Total     int64       `json:"total"`
}
