// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

var _ ExecutionStore = (*executions)(nil)

var Executions ExecutionStore

type ExecutionStore interface {
	// Create records a sandbox execution; text fields are sanitized and truncated to MaxExecutionTextBytes.
	Create(ctx context.Context, options CreateExecutionOptions) (*Execution, error)
	// List returns executions newest first, without the code and output columns.
	List(ctx context.Context, options ListExecutionOptions) ([]*Execution, int64, error)
	// GetByID returns an execution with the given id.
	GetByID(ctx context.Context, id uint) (*Execution, error)
	// Stats returns totals and per-day counts for the last `days` days in the given location.
	Stats(ctx context.Context, days int, location *time.Location) (*ExecutionStats, error)
	// DeleteBefore deletes executions created before t and returns how many were deleted.
	DeleteBefore(ctx context.Context, t time.Time) (int64, error)
}

func NewExecutionStore(db *gorm.DB) ExecutionStore {
	return &executions{db}
}

type executions struct {
	*gorm.DB
}

type ExecutionStatus string

const (
	ExecutionStatusSuccess  ExecutionStatus = "success"
	ExecutionStatusFailed   ExecutionStatus = "failed"
	ExecutionStatusTimeout  ExecutionStatus = "timeout"
	ExecutionStatusRejected ExecutionStatus = "rejected"
	ExecutionStatusError    ExecutionStatus = "error"
)

func (s ExecutionStatus) Valid() bool {
	switch s {
	case ExecutionStatusSuccess, ExecutionStatusFailed, ExecutionStatusTimeout, ExecutionStatusRejected, ExecutionStatusError:
		return true
	}
	return false
}

// MaxExecutionTextBytes caps each stored code and output field.
const MaxExecutionTextBytes = 64 * 1024

type Execution struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`

	SandboxID uint     `gorm:"index;not null"`
	Sandbox   *Sandbox `gorm:"foreignKey:SandboxID;-:migration"`

	Language  string
	Status    ExecutionStatus `gorm:"index;not null"`
	Code      string
	CodeBytes int
	Stdout    string
	Stderr    string
	Error     string
	// Truncated reports whether any text field was cut at MaxExecutionTextBytes.
	Truncated bool
	// Duration is the WASM run time in nanoseconds; zero when the task did not run.
	Duration int64

	ClientAddress string
	UserAgent     string
}

var ErrExecutionNotFound = errors.New("execution does not exist")

type CreateExecutionOptions struct {
	SandboxID     uint
	Language      string
	Status        ExecutionStatus
	Code          string
	CodeBytes     int
	Stdout        string
	Stderr        string
	Error         string
	Duration      int64
	ClientAddress string
	UserAgent     string
}

// sanitizeText makes s storable in PostgreSQL text (valid UTF-8, no NUL) and cuts it to max bytes on a rune boundary.
func sanitizeText(s string, max int) (string, bool) {
	s = strings.ReplaceAll(strings.ToValidUTF8(s, "\uFFFD"), "\x00", "")
	if len(s) <= max {
		return s, false
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut], true
}

func (db *executions) Create(ctx context.Context, options CreateExecutionOptions) (*Execution, error) {
	execution := &Execution{
		SandboxID: options.SandboxID,
		Status:    options.Status,
		CodeBytes: options.CodeBytes,
		Duration:  options.Duration,
	}
	for _, field := range []struct {
		dst *string
		src string
		max int
	}{
		{&execution.Language, options.Language, 32},
		{&execution.Code, options.Code, MaxExecutionTextBytes},
		{&execution.Stdout, options.Stdout, MaxExecutionTextBytes},
		{&execution.Stderr, options.Stderr, MaxExecutionTextBytes},
		{&execution.Error, options.Error, MaxExecutionTextBytes},
		{&execution.ClientAddress, options.ClientAddress, 64},
		{&execution.UserAgent, options.UserAgent, 512},
	} {
		value, truncated := sanitizeText(field.src, field.max)
		*field.dst = value
		if truncated && field.max == MaxExecutionTextBytes {
			execution.Truncated = true
		}
	}

	if err := db.WithContext(ctx).Create(execution).Error; err != nil {
		return nil, errors.Wrap(err, "create")
	}
	return execution, nil
}

// preloadSandbox includes soft-deleted sandboxes so old records keep their names.
func preloadSandbox(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Sandbox", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() })
}

type ListExecutionOptions struct {
	dbutil.Pagination
	SandboxID uint
	Status    ExecutionStatus
}

func (db *executions) List(ctx context.Context, options ListExecutionOptions) ([]*Execution, int64, error) {
	query := db.WithContext(ctx).Model(&Execution{})
	if options.SandboxID != 0 {
		query = query.Where("sandbox_id = ?", options.SandboxID)
	}
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "count")
	}

	var items []*Execution
	limit, offset := options.LimitOffset()
	if err := preloadSandbox(query).Omit("code", "stdout", "stderr").
		Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, errors.Wrap(err, "find")
	}
	return items, total, nil
}

func (db *executions) GetByID(ctx context.Context, id uint) (*Execution, error) {
	var execution Execution
	if err := preloadSandbox(db.WithContext(ctx)).Where("id = ?", id).First(&execution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExecutionNotFound
		}
		return nil, errors.Wrap(err, "first")
	}
	return &execution, nil
}

type ExecutionDailyCount struct {
	// Date is the local day in YYYY-MM-DD.
	Date   string
	Total  int64
	Failed int64
}

type ExecutionStats struct {
	Total     int64
	Last24h   int64
	Failed24h int64
	// Daily has one entry per day, oldest first, including days without executions.
	Daily []ExecutionDailyCount
}

func (db *executions) Stats(ctx context.Context, days int, location *time.Location) (*ExecutionStats, error) {
	now := dbutil.Now().In(location)
	stats := &ExecutionStats{}

	tx := db.WithContext(ctx).Model(&Execution{})
	if err := tx.Count(&stats.Total).Error; err != nil {
		return nil, errors.Wrap(err, "count total")
	}
	since := now.Add(-24 * time.Hour)
	if err := db.WithContext(ctx).Model(&Execution{}).
		Select("count(*), count(*) FILTER (WHERE status <> ?)", ExecutionStatusSuccess).
		Where("created_at >= ?", since).
		Row().Scan(&stats.Last24h, &stats.Failed24h); err != nil {
		return nil, errors.Wrap(err, "count last 24 hours")
	}

	start := time.Date(now.Year(), now.Month(), now.Day()-days+1, 0, 0, 0, 0, location)
	var rows []ExecutionDailyCount
	if err := db.WithContext(ctx).Model(&Execution{}).
		Select(`to_char(created_at AT TIME ZONE ?, 'YYYY-MM-DD') AS date,
			count(*) AS total,
			count(*) FILTER (WHERE status <> ?) AS failed`, location.String(), ExecutionStatusSuccess).
		Where("created_at >= ?", start).
		Group("date").
		Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "count per day")
	}

	byDate := make(map[string]ExecutionDailyCount, len(rows))
	for _, row := range rows {
		byDate[row.Date] = row
	}
	stats.Daily = make([]ExecutionDailyCount, 0, days)
	for day := start; !day.After(now); day = day.AddDate(0, 0, 1) {
		date := day.Format(time.DateOnly)
		count := byDate[date]
		count.Date = date
		stats.Daily = append(stats.Daily, count)
	}
	return stats, nil
}

func (db *executions) DeleteBefore(ctx context.Context, t time.Time) (int64, error) {
	result := db.WithContext(ctx).Where("created_at < ?", t).Delete(&Execution{})
	if result.Error != nil {
		return 0, errors.Wrap(result.Error, "delete")
	}
	return result.RowsAffected, nil
}

// RunExecutionRetention deletes executions older than `days` days now and then hourly until ctx is done; zero days disables it.
func RunExecutionRetention(ctx context.Context, days int) {
	if days <= 0 {
		return
	}
	cleanup := func() {
		deleted, err := Executions.DeleteBefore(ctx, dbutil.Now().AddDate(0, 0, -days))
		if err != nil {
			if ctx.Err() == nil {
				logrus.WithError(err).Error("Failed to delete expired executions")
			}
			return
		}
		if deleted > 0 {
			logrus.WithField("count", deleted).Info("Deleted expired executions")
		}
	}

	cleanup()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanup()
		}
	}
}
