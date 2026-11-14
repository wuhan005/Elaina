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
	Language       []string `json:"language" extensions:"x-nullable"`
	Timeout        int      `json:"timeout"`
	MaxMemory      int64    `json:"maxMemory"`
	InternetAccess bool     `json:"internetAccess"`
} // @name Template

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
	}
}

func ToTemplates(templates []*db.Tpl) []*Template {
	items := make([]*Template, 0, len(templates))
	for _, template := range templates {
		items = append(items, ToTemplate(template))
	}
	return items
}

type ListTemplatesResp struct {
	Templates []*Template `json:"templates"`
	Total     int64       `json:"total"`
} // @name ListTemplatesResp
