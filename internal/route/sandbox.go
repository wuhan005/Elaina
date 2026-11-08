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

type SandboxHandler struct{}

func NewSandboxHandler() *SandboxHandler {
	return &SandboxHandler{}
}

func (h *SandboxHandler) Sandboxer(ctx context.Context) error {
	sandboxID := uint(ctx.ParamInt("id"))

	sandbox, err := db.Sandboxes.GetByID(ctx.Request().Context(), sandboxID)
	if err != nil {
		if errors.Is(err, db.ErrSandboxNotFound) {
			return ctx.Error(http.StatusNotFound, "Sanbox does not exist")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get sandbox")
		return ctx.ServerError()
	}

	ctx.Map(sandbox)
	return nil
}

// List list sandboxes.
// @Summary List sandboxes
// @Tags Sandboxes
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number; values below 1 use 1" default(1)
// @Param pageSize query int false "Page size; values below 1 use 20" default(20)
// @Success 200 {object} dto.Response[dto.ListSandboxesResp]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID listSandboxes
// @Router /api/sandboxes [get]
func (h *SandboxHandler) List(ctx context.Context) error {
	sandboxes, total, err := db.Sandboxes.List(ctx.Request().Context(), db.ListSandboxOptions{
		Pagination: dbutil.Pagination{
			Page:     ctx.QueryInt("page"),
			PageSize: ctx.QueryInt("pageSize"),
		},
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to list sandboxes")
		return ctx.ServerError()
	}

	items := make([]*dto.Sandbox, 0, len(sandboxes))
	for _, sandbox := range sandboxes {
		items = append(items, dto.ToSandbox(sandbox))
	}

	return ctx.Success(dto.ListSandboxesResp{
		Sandboxes: items,
		Total:     total,
	})
}

// Get get a sandbox.
// @Summary Get a sandbox
// @Tags Sandboxes
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[dto.Sandbox]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID getSandbox
// @Router /api/sandbox/{id} [get]
func (h *SandboxHandler) Get(ctx context.Context, sandbox *db.Sandbox) error {
	return ctx.Success(dto.ToSandbox(sandbox))
}

// Create create a sandbox.
// @Summary Create a sandbox
// @Tags Sandboxes
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.CreateSandbox true "Request body"
// @Success 200 {object} dto.Response[dto.Sandbox]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID createSandbox
// @Router /api/sandboxes [post]
func (h *SandboxHandler) Create(ctx context.Context, f form.CreateSandbox) error {
	sandbox, err := db.Sandboxes.Create(ctx.Request().Context(), db.CreateSandboxOptions{
		Name:        f.Name,
		TemplateID:  f.TemplateID,
		Placeholder: f.Placeholder,
		Editable:    f.Editable,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create sandbox")
		return ctx.ServerError()
	}

	return ctx.Success(dto.ToSandbox(sandbox))
}

// Update update a sandbox.
// @Summary Update a sandbox
// @Tags Sandboxes
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Param data body form.UpdateSandbox true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID updateSandbox
// @Router /api/sandbox/{id} [put]
func (h *SandboxHandler) Update(ctx context.Context, sandbox *db.Sandbox, f form.UpdateSandbox) error {
	sandboxID := sandbox.ID

	if err := db.Sandboxes.Update(ctx.Request().Context(), sandboxID, db.UpdateSandboxOptions{
		Name:        f.Name,
		TemplateID:  f.TemplateID,
		Placeholder: f.Placeholder,
		Editable:    f.Editable,
	}); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update sandbox")
		return ctx.ServerError()
	}

	return ctx.Success("SandBox updated successfully")
}

// Delete delete a sandbox.
// @Summary Delete a sandbox
// @Tags Sandboxes
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[string]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID deleteSandbox
// @Router /api/sandbox/{id} [delete]
func (h *SandboxHandler) Delete(ctx context.Context, sandbox *db.Sandbox) error {
	sandboxID := sandbox.ID

	if err := db.Sandboxes.Delete(ctx.Request().Context(), sandboxID); err != nil {
		if errors.Is(err, db.ErrSandboxNotFound) {
			return ctx.Error(http.StatusNotFound, "Sandbox does not exist")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete sandbox")
		return ctx.ServerError()
	}
	return ctx.Success("Sandbox deleted successfully")
}
