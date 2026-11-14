// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

import (
	"time"

	"github.com/wuhan005/Elaina/internal/db"
)

type ExecutionSandbox struct {
	ID      uint   `json:"id"`
	UID     string `json:"uid"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
} // @name ExecutionSandbox

type ExecutionSummary struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	SandboxID uint              `json:"sandboxID"`
	Sandbox   *ExecutionSandbox `json:"sandbox" extensions:"x-nullable"`
	Language  string            `json:"language"`
	Status    string            `json:"status" enums:"success,failed,timeout,rejected,error"`
	Error     string            `json:"error"`
	// WASM run time in nanoseconds; 0 when the task did not run.
	Duration int64 `json:"duration"`
	// Size of the submitted source code in bytes.
	CodeBytes     int    `json:"codeBytes"`
	ClientAddress string `json:"clientAddress"`
	UserAgent     string `json:"userAgent"`
} // @name ExecutionSummary

type ExecutionDetail struct {
	ExecutionSummary

	// Empty for rejected requests.
	Code   string `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	// True when code or output was cut to 64 KiB.
	Truncated bool `json:"truncated"`
} // @name ExecutionDetail

func toExecutionSummary(execution *db.Execution) ExecutionSummary {
	summary := ExecutionSummary{
		ID:            execution.ID,
		CreatedAt:     execution.CreatedAt,
		SandboxID:     execution.SandboxID,
		Language:      execution.Language,
		Status:        string(execution.Status),
		Error:         execution.Error,
		Duration:      execution.Duration,
		CodeBytes:     execution.CodeBytes,
		ClientAddress: execution.ClientAddress,
		UserAgent:     execution.UserAgent,
	}
	if sandbox := execution.Sandbox; sandbox != nil {
		summary.Sandbox = &ExecutionSandbox{
			ID:      sandbox.ID,
			UID:     sandbox.UID,
			Name:    sandbox.Name,
			Deleted: sandbox.DeletedAt.Valid,
		}
	}
	return summary
}

func ToExecutionSummaries(executions []*db.Execution) []ExecutionSummary {
	items := make([]ExecutionSummary, 0, len(executions))
	for _, execution := range executions {
		items = append(items, toExecutionSummary(execution))
	}
	return items
}

func ToExecutionDetail(execution *db.Execution) *ExecutionDetail {
	if execution == nil {
		return nil
	}
	return &ExecutionDetail{
		ExecutionSummary: toExecutionSummary(execution),
		Code:             execution.Code,
		Stdout:           execution.Stdout,
		Stderr:           execution.Stderr,
		Truncated:        execution.Truncated,
	}
}

type ListExecutionsResp struct {
	Executions []ExecutionSummary `json:"executions"`
	Total      int64              `json:"total"`
} // @name ListExecutionsResp

type ExecutionDailyCount struct {
	// Local day in YYYY-MM-DD.
	Date   string `json:"date"`
	Total  int64  `json:"total"`
	Failed int64  `json:"failed"`
} // @name ExecutionDailyCount

type ExecutionStats struct {
	Total     int64 `json:"total"`
	Last24h   int64 `json:"last24h"`
	Failed24h int64 `json:"failed24h"`
	// One entry per day, oldest first.
	Daily []ExecutionDailyCount `json:"daily"`
} // @name ExecutionStats

func ToExecutionStats(stats *db.ExecutionStats) *ExecutionStats {
	daily := make([]ExecutionDailyCount, 0, len(stats.Daily))
	for _, day := range stats.Daily {
		daily = append(daily, ExecutionDailyCount{Date: day.Date, Total: day.Total, Failed: day.Failed})
	}
	return &ExecutionStats{
		Total:     stats.Total,
		Last24h:   stats.Last24h,
		Failed24h: stats.Failed24h,
		Daily:     daily,
	}
}
