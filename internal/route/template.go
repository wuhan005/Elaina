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
	"github.com/wuhan005/Elaina/internal/dto"
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
			return ctx.Error(http.StatusNotFound, "template::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get template")
		return ctx.ServerError()
	}

	ctx.Map(tpl)
	return nil
}

// All get all templates.
// @Summary Get all templates
// @Tags Templates
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[[]dto.Template]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID allTemplates
// @Router /api/templates/all [get]
func (h *TemplateHandler) All(ctx context.Context) error {
	templates, err := db.Tpls.All(ctx.Request().Context())
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get all templates")
		return ctx.ServerError()
	}

	return ctx.Success(dto.ToTemplates(templates))
}

// List list templates.
// @Summary List templates
// @Tags Templates
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number; values below 1 use 1" default(1)
// @Param pageSize query int false "Page size; values below 1 use 20" default(20)
// @Success 200 {object} dto.Response[dto.ListTemplatesResp]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID listTemplates
// @Router /api/templates [get]
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

	return ctx.Success(dto.ListTemplatesResp{
		Templates: dto.ToTemplates(templates),
		Total:     total,
	})
}

// Get get a template.
// @Summary Get a template
// @Tags Templates
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[dto.Template]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID getTemplate
// @Router /api/template/{id} [get]
func (h *TemplateHandler) Get(ctx context.Context, template *db.Tpl) error {
	return ctx.Success(dto.ToTemplate(template))
}

// Create create a template.
// @Summary Create a template
// @Tags Templates
// @Description Zero timeout uses 10 seconds. Zero memory inherits app.wasm_limits.max_memory_mib; nonzero memory is capped by that ceiling for each stage.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.CreateTemplate true "Request body"
// @Success 200 {object} dto.Response[dto.Template]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID createTemplate
// @Router /api/templates [post]
func (h *TemplateHandler) Create(ctx context.Context, f form.CreateTemplate) error {
	template, err := db.Tpls.Create(ctx.Request().Context(), db.CreateTplOptions{
		Name:           f.Name,
		Language:       f.Language,
		Timeout:        f.Timeout,
		MaxMemory:      f.MaxMemory,
		InternetAccess: f.InternetAccess,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create template")
		return ctx.ServerError()
	}

	return ctx.Success(dto.ToTemplate(template))
}

// Update update a template.
// @Summary Update a template
// @Tags Templates
// @Description Zero timeout uses 10 seconds. Zero memory inherits app.wasm_limits.max_memory_mib; nonzero memory is capped by that ceiling for each stage.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Param data body form.UpdateTemplate true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID updateTemplate
// @Router /api/template/{id} [put]
func (h *TemplateHandler) Update(ctx context.Context, template *db.Tpl, f form.UpdateTemplate) error {
	if err := db.Tpls.Update(ctx.Request().Context(), template.ID, db.UpdateTplOptions{
		Name:           f.Name,
		Language:       f.Language,
		Timeout:        f.Timeout,
		MaxMemory:      f.MaxMemory,
		InternetAccess: f.InternetAccess,
	}); err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			return ctx.Error(http.StatusNotFound, "template::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update template")
		return ctx.ServerError()
	}
	return ctx.Success(ctx.Tr("template::updated"))
}

// Delete delete a template.
// @Summary Delete a template
// @Tags Templates
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[string]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID deleteTemplate
// @Router /api/template/{id} [delete]
func (h *TemplateHandler) Delete(ctx context.Context, template *db.Tpl) error {
	if err := db.Tpls.Delete(ctx.Request().Context(), template.ID); err != nil {
		if errors.Is(err, db.ErrTemplateNotFound) {
			return ctx.Error(http.StatusNotFound, "template::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete template")
		return ctx.ServerError()
	}
	return ctx.Success(ctx.Tr("template::deleted"))
}
