// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/config"
)

func TestWASMAdmission(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxConcurrentTasks = 1
	config.App.WASMLimits.MaxQueuedTasks = 1
	config.App.WASMLimits.MaxTasksPerClient = 1
	a := &wasmAdmission{clients: map[string]int{}}
	ctx := context.Background()
	queued := func(n int) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for {
			a.mu.Lock()
			length := len(a.queue)
			a.mu.Unlock()
			if length == n {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("queue length %d, want %d", length, n)
			}
			time.Sleep(time.Millisecond)
		}
	}

	releaseA, err := a.acquire(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.acquire(ctx, "a"); !errors.Is(err, ErrClientTaskLimit) {
		t.Fatalf("one client held two tasks: %v", err)
	}

	acquiredB := make(chan func())
	go func() {
		release, err := a.acquire(ctx, "b")
		if err != nil {
			t.Error(err)
		}
		acquiredB <- release
	}()
	queued(1)
	if _, err := a.acquire(ctx, "c"); !errors.Is(err, ErrTaskLimit) {
		t.Fatalf("full queue accepted a task: %v", err)
	}

	releaseA()
	releaseB := <-acquiredB
	if a.running() != 1 {
		t.Fatalf("slot was not handed to the queued client: running=%d", a.running())
	}

	cancelled, cancel := context.WithCancel(ctx)
	done := make(chan error)
	go func() {
		_, err := a.acquire(cancelled, "d")
		done <- err
	}()
	queued(1)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled waiter: %v", err)
	}
	queued(0)

	releaseB()
	if a.running() != 0 || len(a.clients) != 0 {
		t.Fatalf("admission leaked state: running=%d clients=%v", a.running(), a.clients)
	}

	config.App.WASMLimits.MaxConcurrentTasks = 2
	for range 2 {
		release, err := a.acquire(ctx, "")
		if err != nil {
			t.Fatalf("anonymous internal task was limited per client: %v", err)
		}
		defer release()
	}
}
