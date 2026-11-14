// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"time"
	_ "time/tzdata" // Resolve ?timezone= on hosts without system time zone data.

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dbutil"
	"github.com/wuhan005/Elaina/internal/dto"
)

const (
	defaultStatsDays = 14
	maxStatsDays     = 90
)

type ExecutionHandler struct{}

func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{}
}

func (h *ExecutionHandler) Executioner(ctx context.Context) error {
	execution, err := db.Executions.GetByID(ctx.Request().Context(), uint(ctx.ParamInt("id")))
	if err != nil {
		if errors.Is(err, db.ErrExecutionNotFound) {
			return ctx.Error(http.StatusNotFound, "execution::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get execution")
		return ctx.ServerError()
	}

	ctx.Map(execution)
	return nil
}

// List list sandbox executions.
// @Summary List sandbox executions
// @Tags Executions
// @Description Newest first. Code and output are only included in getExecution.
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number; values below 1 use 1" default(1)
// @Param pageSize query int false "Page size; values below 1 use 20" default(20)
// @Param sandboxID query int false "Only executions of this sandbox"
// @Param status query string false "Only executions with this status" Enums(success, failed, timeout, rejected, error)
// @Success 200 {object} dto.Response[dto.ListExecutionsResp]
// @Failure 400 {object} dto.ErrorResponse "Invalid status"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID listExecutions
// @Router /api/executions [get]
func (h *ExecutionHandler) List(ctx context.Context) error {
	status := db.ExecutionStatus(ctx.Query("status"))
	if status != "" && !status.Valid() {
		return ctx.Error(http.StatusBadRequest, "execution::invalid_status", status)
	}

	executions, total, err := db.Executions.List(ctx.Request().Context(), db.ListExecutionOptions{
		Pagination: dbutil.Pagination{
			Page:     ctx.QueryInt("page"),
			PageSize: ctx.QueryInt("pageSize"),
		},
		SandboxID: uint(max(ctx.QueryInt("sandboxID"), 0)),
		Status:    status,
	})
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to list executions")
		return ctx.ServerError()
	}

	return ctx.Success(dto.ListExecutionsResp{
		Executions: dto.ToExecutionSummaries(executions),
		Total:      total,
	})
}

// Get get a sandbox execution.
// @Summary Get a sandbox execution
// @Tags Executions
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[dto.ExecutionDetail]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID getExecution
// @Router /api/execution/{id} [get]
func (h *ExecutionHandler) Get(ctx context.Context, execution *db.Execution) error {
	return ctx.Success(dto.ToExecutionDetail(execution))
}

// Stats get execution statistics.
// @Summary Get execution statistics
// @Tags Executions
// @Produce json
// @Security SessionAuth
// @Param days query int false "Number of days in daily, including today (1-90)" default(14)
// @Param timezone query string false "IANA time zone used to split days" default(UTC)
// @Success 200 {object} dto.Response[dto.ExecutionStats]
// @Failure 400 {object} dto.ErrorResponse "Invalid time zone or days"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID executionStats
// @Router /api/executions/stats [get]
func (h *ExecutionHandler) Stats(ctx context.Context) error {
	days := ctx.QueryInt("days")
	if days == 0 {
		days = defaultStatsDays
	}
	if days < 1 || days > maxStatsDays {
		return ctx.Error(http.StatusBadRequest, "execution::invalid_days", maxStatsDays)
	}

	location := time.UTC
	if name := ctx.Query("timezone"); name != "" {
		loaded, err := time.LoadLocation(name)
		if err != nil || name == "Local" {
			return ctx.Error(http.StatusBadRequest, "execution::unknown_timezone", name)
		}
		location = loaded
	}

	stats, err := db.Executions.Stats(ctx.Request().Context(), days, location)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get execution stats")
		return ctx.ServerError()
	}
	return ctx.Success(dto.ToExecutionStats(stats))
}
