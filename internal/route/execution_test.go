// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/runtime"
)

// fakeExecutions is an in-memory ExecutionStore that remembers its last call.
type fakeExecutions struct {
	mu         sync.Mutex
	created    []db.CreateExecutionOptions
	listed     db.ListExecutionOptions
	statsDays  int
	statsZone  string
	executions map[uint]*db.Execution
}

func useFakeExecutions(t *testing.T) *fakeExecutions {
	t.Helper()
	store := &fakeExecutions{executions: map[uint]*db.Execution{}}
	previous := db.Executions
	db.Executions = store
	t.Cleanup(func() { db.Executions = previous })
	return store
}

func (s *fakeExecutions) Create(_ context.Context, options db.CreateExecutionOptions) (*db.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created = append(s.created, options)
	return &db.Execution{}, nil
}

func (s *fakeExecutions) last() db.CreateExecutionOptions {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.created) == 0 {
		return db.CreateExecutionOptions{}
	}
	return s.created[len(s.created)-1]
}

func (s *fakeExecutions) List(_ context.Context, options db.ListExecutionOptions) ([]*db.Execution, int64, error) {
	s.listed = options
	return []*db.Execution{}, 0, nil
}

func (s *fakeExecutions) GetByID(_ context.Context, id uint) (*db.Execution, error) {
	if execution, ok := s.executions[id]; ok {
		return execution, nil
	}
	return nil, db.ErrExecutionNotFound
}

func (s *fakeExecutions) Stats(_ context.Context, days int, location *time.Location) (*db.ExecutionStats, error) {
	s.statsDays, s.statsZone = days, location.String()
	return &db.ExecutionStats{Daily: []db.ExecutionDailyCount{}}, nil
}

func (s *fakeExecutions) DeleteBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestExecutionRoutes(t *testing.T) {
	users := useFakeUsers(t)
	executions := useFakeExecutions(t)
	executions.executions[7] = &db.Execution{
		ID:        7,
		SandboxID: 3,
		Sandbox:   &db.Sandbox{UID: "abc", Name: "Hello"},
		Status:    db.ExecutionStatusSuccess,
		Code:      "print(1)",
		Stdout:    "1\n",
	}
	createTestUser(t, users, "admin", "admin-password")
	client := newTestClient(t)
	_, token, _ := client.signIn("admin", "admin-password")

	if status, _ := client.do(http.MethodGet, "/api/executions", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("anonymous list: expected 401, got %d", status)
	}

	for _, test := range []struct {
		name, path string
		status     int
	}{
		{"list", "/api/executions?page=2&pageSize=5&sandboxID=3&status=failed", http.StatusOK},
		{"invalid status", "/api/executions?status=unknown", http.StatusBadRequest},
		{"stats", "/api/executions/stats?days=7&timezone=Asia/Shanghai", http.StatusOK},
		{"default stats", "/api/executions/stats", http.StatusOK},
		{"unknown time zone", "/api/executions/stats?timezone=Mars/Base", http.StatusBadRequest},
		{"local time zone", "/api/executions/stats?timezone=Local", http.StatusBadRequest},
		{"too many days", "/api/executions/stats?days=91", http.StatusBadRequest},
		{"detail", "/api/execution/7", http.StatusOK},
		{"missing detail", "/api/execution/8", http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			if status, _ := client.do(http.MethodGet, test.path, token, nil); status != test.status {
				t.Fatalf("expected status %d, got %d", test.status, status)
			}
		})
	}

	if listed := executions.listed; listed.Page != 2 || listed.PageSize != 5 || listed.SandboxID != 3 || listed.Status != db.ExecutionStatusFailed {
		t.Fatalf("unexpected list options: %+v", listed)
	}
	if executions.statsDays != defaultStatsDays || executions.statsZone != "UTC" {
		t.Fatalf("unexpected default stats options: %d %s", executions.statsDays, executions.statsZone)
	}

	_, detail := client.do(http.MethodGet, "/api/execution/7", token, nil)
	data := detail["data"].(map[string]any)
	if data["code"] != "print(1)" || data["stdout"] != "1\n" || data["sandbox"].(map[string]any)["name"] != "Hello" {
		t.Fatalf("unexpected detail: %v", data)
	}
}

func TestExecutionRecord(t *testing.T) {
	executions := useFakeExecutions(t)
	request := httptest.NewRequest(http.MethodPost, "/r/abc/execute", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set("User-Agent", "test-agent")
	record := executionRecord{
		request:  request,
		sandbox:  &db.Sandbox{UID: "abc"},
		language: "python",
		code:     "print(1)",
	}
	record.sandbox.ID = 3

	record.save(db.ExecutionStatusSuccess, &runtime.ExecOutput{Stdout: "1\n", Duration: 42}, nil)
	record.save(db.ExecutionStatusRejected, nil, errors.New("too many tasks"))

	if len(executions.created) != 2 {
		t.Fatalf("expected 2 records, got %d", len(executions.created))
	}
	success, rejected := executions.created[0], executions.created[1]
	if success.SandboxID != 3 || success.Code != "print(1)" || success.Stdout != "1\n" || success.Duration != 42 ||
		success.ClientAddress != "192.0.2.10" || success.UserAgent != "test-agent" || success.CodeBytes != 8 {
		t.Fatalf("unexpected success record: %+v", success)
	}
	if rejected.Code != "" || rejected.CodeBytes != 8 || !strings.Contains(rejected.Error, "too many tasks") {
		t.Fatalf("rejected records must keep only the code size: %+v", rejected)
	}

	// Records keep the full IPv6 address, not the /64 rate-limit key.
	request.RemoteAddr = "[2001:db8::1234]:443"
	record.save(db.ExecutionStatusSuccess, &runtime.ExecOutput{}, nil)
	if got := executions.last().ClientAddress; got != "2001:db8::1234" {
		t.Fatalf("expected the full IPv6 address, got %q", got)
	}
}
