//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
)

func TestWASMTaskAdmission(t *testing.T) {
	configureWASMCacheTest(t)
	config.App.WASMLimits.MaxConcurrentTasks = 2
	installWASMCacheTestModule(t, "python.wasm", wasmTestModule([]byte{0x03, 0x40, 0x0c, 0, 0x0b}))
	installWASMCacheTestModule(t, "c/clang.wasm", wasmTestModule([]byte{0x03, 0x40, 0x0c, 0, 0x0b}))
	installWASMCacheTestModule(t, "javascript.wasm", wasmTestModule([]byte{0x00}))
	if err := Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 2)
	for _, language := range []string{"python", "c"} {
		go func() {
			task, err := NewWASMTask(ctx, NewTaskOptions{
				Language: language,
				Template: &db.Tpl{},
			})
			if err == nil {
				_, err = task.Run(ctx)
			}
			done <- err
		}()
	}
	defer func() {
		cancel()
		for range 2 {
			if err := <-done; !errors.Is(err, context.Canceled) {
				t.Errorf("cancel running task: %v", err)
			}
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for activeWASMTasks.Load() != 2 {
		if time.Now().After(deadline) {
			t.Fatal("tasks did not acquire their slots")
		}
		time.Sleep(time.Millisecond)
	}

	other, err := NewWASMTask(ctx, NewTaskOptions{
		Language: "javascript",
		Template: &db.Tpl{},
	})
	if err != nil {
		t.Fatal(err)
	}
	for range 10 {
		if _, err := other.Run(context.Background()); !errors.Is(err, ErrTaskLimit) {
			t.Fatalf("global admission did not reject a different template/language: %v", err)
		}
	}
	cancel()
	for activeWASMTasks.Load() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("cancellation leaked a task slot")
		}
		time.Sleep(time.Millisecond)
	}
	for range 3 {
		output, err := other.Run(context.Background())
		if err != nil || output == nil || !strings.Contains(output.Error, "unreachable") || activeWASMTasks.Load() != 0 {
			t.Fatalf("failed task leaked a slot: %+v, %v", output, err)
		}
	}
}

func TestWASMMemoryCeiling(t *testing.T) {
	configureWASMCacheTest(t)
	config.App.WASMLimits.MaxMemoryMiB = 1
	// Growing to 17 pages must fail at the one-MiB server ceiling.
	installWASMCacheTestModule(t, "python.wasm", wasmTestModule([]byte{0x41, 16, 0x40, 0, 0x41, 0x7f, 0x47, 0x04, 0x40, 0x00, 0x0b}))
	if err := Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, memory := range []int64{0, 1, 512, 4096} {
		task, err := NewWASMTask(context.Background(), NewTaskOptions{
			Language: "python",
			Template: &db.Tpl{
				MaxMemory: memory,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		output, err := task.Run(context.Background())
		if err != nil || output == nil || output.Error != "" {
			t.Fatalf("memory %d bypassed ceiling: %+v, %v", memory, output, err)
		}
	}
}

func TestWASMSourceLimit(t *testing.T) {
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxSourceBytes = 6
	for _, code := range []string{"1234567", "中文a"} {
		_, err := NewWASMTask(context.Background(), NewTaskOptions{
			Language: "python",
			Template: &db.Tpl{},
			Code:     []byte(code),
		})
		if !errors.Is(err, ErrSourceLimit) {
			t.Fatalf("oversized source accepted: %v", err)
		}
	}
	task, err := NewWASMTask(context.Background(), NewTaskOptions{
		Language: "python",
		Template: &db.Tpl{},
		Code:     []byte("中文"),
	})
	if err != nil {
		t.Fatal(err)
	}
	task.code = append(task.code, 'x')
	if _, err := task.Run(context.Background()); !errors.Is(err, ErrSourceLimit) || activeWASMTasks.Load() != 0 {
		t.Fatalf("runtime entry bypassed source limit: %v", err)
	}
}
