// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dbutil"
	"github.com/wuhan005/Elaina/internal/form"
)

type TemplateHandler struct{}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

func (h *TemplateHandler) Templater(ctx context.Context) error {
	templateID := uint(ctx.ParamInt("id"))
	tpl, err := db.Tpls.GetByID(ctx.Request().Context(), templateID)
	if err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			return ctx.Error(http.StatusNotFound, "Template dose not exist")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get template")
		return ctx.ServerError()
	}

	ctx.Map(tpl)
	return nil
}

func (h *TemplateHandler) All(ctx context.Context) error {
	templates, err := db.Tpls.All(ctx.Request().Context())
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get all templates")
		return ctx.ServerError()
	}
	return ctx.Success(templates)
}

func (h *TemplateHandler) List(ctx context.Context) error {
	templates, total, err := db.Tpls.List(ctx.Request().Context(), db.ListTplOptions{
		Pagination: dbutil.Pagination{
			Page:     ctx.QueryInt("page"),
			PageSize: ctx.QueryInt("pageSize"),
		},
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to list templates")
		return ctx.ServerError()
	}

	return ctx.Success(map[string]interface{}{
		"templates": templates,
		"total":     total,
	})
}

func (h *TemplateHandler) Get(ctx context.Context, template *db.Tpl) error {
	return ctx.Success(template)
}

func (h *TemplateHandler) Create(ctx context.Context, f form.CreateTemplate) error {
	template, err := db.Tpls.Create(ctx.Request().Context(), db.CreateTplOptions{
		Name:              f.Name,
		Language:          f.Language,
		Timeout:           f.Timeout,
		MaxCPUs:           f.MaxCPUs,
		MaxMemory:         f.MaxMemory,
		InternetAccess:    f.InternetAccess,
		DNS:               f.DNS,
		MaxContainer:      f.MaxContainer,
		MaxContainerPerIP: f.MaxContainerPerIP,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create template")
		return ctx.ServerError()
	}
	return ctx.Success(template)
}

func (h *TemplateHandler) Update(ctx context.Context, template *db.Tpl, f form.UpdateTemplate) error {
	templateID := template.ID

	if err := db.Tpls.Update(ctx.Request().Context(), templateID, db.UpdateTplOptions{
		Name:              f.Name,
		Language:          f.Language,
		Timeout:           f.Timeout,
		MaxCPUs:           f.MaxCPUs,
		MaxMemory:         f.MaxMemory,
		InternetAccess:    f.InternetAccess,
		DNS:               f.DNS,
		MaxContainer:      f.MaxContainer,
		MaxContainerPerIP: f.MaxContainerPerIP,
	}); err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			return ctx.Error(http.StatusNotFound, "Template dose not exist")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update template")
		return ctx.ServerError()
	}
	return ctx.Success("Template updated successfully")
}

func (h *TemplateHandler) Delete(ctx context.Context, template *db.Tpl) error {
	templateID := template.ID

	if err := db.Tpls.Delete(ctx.Request().Context(), templateID); err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			return ctx.Error(http.StatusNotFound, "Template dose not exist")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete template")
		return ctx.ServerError()
	}
	return ctx.Success("Template deleted successfully")
}
