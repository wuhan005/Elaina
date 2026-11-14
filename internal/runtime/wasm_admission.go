// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"slices"
	"sync"

	"github.com/pkg/errors"

	"github.com/wuhan005/Elaina/internal/config"
)

var (
	ErrTaskLimit       = errors.New("WASM task limit reached; try again later")
	ErrClientTaskLimit = errors.New("too many WASM tasks from this client; wait for the previous ones to finish")
)

// wasmAdmission manages global execution slots with a bounded FIFO queue, and caps running plus queued tasks per
// client so that a single source cannot monopolize the slots.
type wasmAdmission struct {
	mu      sync.Mutex
	active  int
	queue   []*wasmWaiter
	clients map[string]int
}

type wasmWaiter struct {
	ready chan struct{}
}

var taskAdmission = &wasmAdmission{clients: map[string]int{}}

// acquire obtains a slot for client; on success the returned release must be called exactly once.
// An empty client is exempt from the per-client limit.
func (a *wasmAdmission) acquire(ctx context.Context, client string) (func(), error) {
	limits := config.App.WASMLimits
	release := func() { a.release(client) }

	a.mu.Lock()
	if client != "" && a.clients[client] >= limits.MaxTasksPerClient {
		a.mu.Unlock()
		return nil, ErrClientTaskLimit
	}
	if a.active < limits.MaxConcurrentTasks && len(a.queue) == 0 {
		a.active++
		a.clients[client]++
		a.mu.Unlock()
		return release, nil
	}
	if len(a.queue) >= limits.MaxQueuedTasks {
		a.mu.Unlock()
		return nil, ErrTaskLimit
	}
	waiter := &wasmWaiter{ready: make(chan struct{})}
	a.queue = append(a.queue, waiter)
	a.clients[client]++
	a.mu.Unlock()

	select {
	case <-waiter.ready:
		return release, nil
	case <-ctx.Done():
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	select {
	case <-waiter.ready:
		// The slot was handed over while cancelling; pass it on to the next waiter.
		a.releaseLocked(client)
	default:
		a.queue = slices.DeleteFunc(a.queue, func(w *wasmWaiter) bool { return w == waiter })
		a.leaveLocked(client)
	}
	return nil, ctx.Err()
}

func (a *wasmAdmission) release(client string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releaseLocked(client)
}

func (a *wasmAdmission) releaseLocked(client string) {
	a.leaveLocked(client)
	if len(a.queue) > 0 && a.active <= config.App.WASMLimits.MaxConcurrentTasks {
		next := a.queue[0]
		a.queue = slices.Delete(a.queue, 0, 1)
		close(next.ready)
		return
	}
	a.active--
}

func (a *wasmAdmission) leaveLocked(client string) {
	if a.clients[client]--; a.clients[client] <= 0 {
		delete(a.clients, client)
	}
}

// running returns the number of tasks holding a slot.
func (a *wasmAdmission) running() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.active
}
