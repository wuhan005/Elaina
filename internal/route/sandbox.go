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

	return ctx.Success(map[string]interface{}{
		"sandboxes": sandboxes,
		"total":     total,
	})
}

func (h *SandboxHandler) Get(ctx context.Context, sandbox *db.Sandbox) error {
	return ctx.Success(sandbox)
}

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

	return ctx.Success(sandbox)
}

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
