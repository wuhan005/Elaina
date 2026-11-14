// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestSanitizeText(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
		max               int
		truncated         bool
	}{
		{"unchanged", "hello", "hello", 10, false},
		{"NUL removed", "a\x00b", "ab", 10, false},
		{"invalid UTF-8 replaced", "a\xffb", "a\uFFFDb", 10, false},
		{"cut to max", "abcdef", "abc", 3, true},
		{"cut on rune boundary", "ab中文", "ab", 4, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, truncated := sanitizeText(test.input, test.max)
			if got != test.want || truncated != test.truncated || !utf8.ValidString(got) {
				t.Fatalf("sanitizeText() = %q, %v; want %q, %v", got, truncated, test.want, test.truncated)
			}
		})
	}
}

func TestExecutions(t *testing.T) {
	database := newTestDatabase(t)
	ctx := context.Background()

	tpl, err := Tpls.Create(ctx, CreateTplOptions{Name: "WASM", Language: []string{"python"}})
	if err != nil {
		t.Fatal(err)
	}
	sandbox, err := Sandboxes.Create(ctx, CreateSandboxOptions{Name: "Hello", TemplateID: tpl.ID})
	if err != nil {
		t.Fatal(err)
	}
	other, err := Sandboxes.Create(ctx, CreateSandboxOptions{Name: "Other", TemplateID: tpl.ID})
	if err != nil {
		t.Fatal(err)
	}

	long := strings.Repeat("x", MaxExecutionTextBytes+10)
	first, err := Executions.Create(ctx, CreateExecutionOptions{
		SandboxID: sandbox.ID,
		Language:  "python",
		Status:    ExecutionStatusSuccess,
		Code:      "print(1)",
		CodeBytes: 8,
		Stdout:    long,
		Stderr:    "a\x00b",
		Duration:  1234,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !first.Truncated || len(first.Stdout) != MaxExecutionTextBytes || first.Stderr != "ab" {
		t.Fatalf("text was not sanitized: truncated=%v stdout=%d stderr=%q", first.Truncated, len(first.Stdout), first.Stderr)
	}
	for _, options := range []CreateExecutionOptions{
		{SandboxID: sandbox.ID, Status: ExecutionStatusFailed, Error: "exit status 1"},
		{SandboxID: other.ID, Status: ExecutionStatusRejected, CodeBytes: 99},
	} {
		if _, err := Executions.Create(ctx, options); err != nil {
			t.Fatal(err)
		}
	}

	items, total, err := Executions.List(ctx, ListExecutionOptions{SandboxID: sandbox.ID})
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("unexpected list: %d %d %v", total, len(items), err)
	}
	if items[0].Status != ExecutionStatusFailed || items[1].Code != "" || items[1].Stdout != "" || items[1].Sandbox == nil || items[1].Sandbox.Name != "Hello" {
		t.Fatalf("list must be newest first, preload the sandbox and omit code/output: %+v", items[1])
	}
	if _, total, _ := Executions.List(ctx, ListExecutionOptions{Status: ExecutionStatusRejected}); total != 1 {
		t.Fatalf("expected 1 rejected execution, got %d", total)
	}

	// Records of deleted sandboxes keep the sandbox name.
	if err := Sandboxes.Delete(ctx, sandbox.ID); err != nil {
		t.Fatal(err)
	}
	got, err := Executions.GetByID(ctx, first.ID)
	if err != nil || got.Code != "print(1)" || got.Sandbox == nil || got.Sandbox.Name != "Hello" || !got.Sandbox.DeletedAt.Valid {
		t.Fatalf("unexpected execution: %+v %v", got, err)
	}
	if _, err := Executions.GetByID(ctx, 0); !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("expected ErrExecutionNotFound, got %v", err)
	}

	// Move the failed execution to three days ago, 23:30 in Shanghai.
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().In(location)
	old := time.Date(now.Year(), now.Month(), now.Day()-3, 23, 30, 0, 0, location)
	if err := database.Model(&Execution{}).Where("status = ?", ExecutionStatusFailed).Update("created_at", old).Error; err != nil {
		t.Fatal(err)
	}

	stats, err := Executions.Stats(ctx, 7, location)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 3 || stats.Last24h != 2 || stats.Failed24h != 1 || len(stats.Daily) != 7 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	byDate := map[string]ExecutionDailyCount{}
	for _, day := range stats.Daily {
		byDate[day.Date] = day
	}
	if day := byDate[old.Format(time.DateOnly)]; day.Total != 1 || day.Failed != 1 {
		t.Fatalf("expected the old execution on %s in local time, got %+v", old.Format(time.DateOnly), stats.Daily)
	}
	if day := byDate[now.Format(time.DateOnly)]; day.Total != 2 || day.Failed != 1 {
		t.Fatalf("unexpected count for today: %+v", day)
	}
	if stats.Daily[6].Date != now.Format(time.DateOnly) {
		t.Fatalf("daily must end today, got %s", stats.Daily[6].Date)
	}

	deleted, err := Executions.DeleteBefore(ctx, now.Add(-24*time.Hour))
	if err != nil || deleted != 1 {
		t.Fatalf("expected 1 deleted execution, got %d (%v)", deleted, err)
	}
}
