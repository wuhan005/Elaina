//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
)

func wasmTestModule(body []byte) []byte {
	// A module with one-page memory, a mutable zero-initialized global, and _start.
	module := []byte{
		0, 'a', 's', 'm', 1, 0, 0, 0,
		1, 4, 1, 0x60, 0, 0,
		3, 2, 1, 0,
		5, 3, 1, 0, 1,
		6, 6, 1, 0x7f, 1, 0x41, 0, 0x0b,
		7, 10, 1, 6, '_', 's', 't', 'a', 'r', 't', 0, 0,
	}
	module = append(module, 10, byte(len(body)+4), 1, byte(len(body)+2), 0)
	module = append(module, body...)
	return append(module, 0x0b)
}

func configureWASMCacheTest(t *testing.T) {
	t.Helper()
	previous := config.App
	t.Cleanup(func() {
		_ = Close(context.Background())
		config.App = previous
	})
	config.App.WASMRuntimeDir = t.TempDir()
	config.App.WASMCacheDir = filepath.Join(t.TempDir(), "cache")
}

func installWASMCacheTestModule(t *testing.T, name string, data []byte) {
	t.Helper()
	path := filepath.Join(config.App.WASMRuntimeDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestWASMCacheWarmup(t *testing.T) {
	configureWASMCacheTest(t)
	// An unreachable entry point proves that warmup compiles without running the module.
	for _, name := range []string{"python.wasm", "php.wasm", "javascript.wasm", "go/compile.wasm", "go/link.wasm", "c/clang.wasm", "c/wasm-ld.wasm"} {
		installWASMCacheTestModule(t, name, wasmTestModule([]byte{0x00}))
	}
	if err := Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	cache := compilationCache
	if err := Init(context.Background()); err == nil || compilationCache != cache {
		t.Fatal("initializing twice replaced the live process cache")
	}

	// Make disk cache reads and writes fail; a retained in-memory entry must still work.
	if err := os.RemoveAll(config.App.WASMCacheDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.App.WASMCacheDir, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		task, err := NewWASMTask(context.Background(), NewTaskOptions{
			Language: "python",
			Template: &db.Tpl{
				MaxMemory: 1,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		output, err := task.Run(context.Background())
		if err != nil || output == nil || !strings.Contains(output.Error, "unreachable") {
			t.Fatalf("warm machine code was not retained after task %d: %+v, %v", i, output, err)
		}
	}
}

func TestWASMCacheInitializationFailure(t *testing.T) {
	configureWASMCacheTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Init(ctx); !errors.Is(err, context.Canceled) || compilationCache != nil {
		t.Fatalf("cancelled initialization published a cache: %v", err)
	}

	installWASMCacheTestModule(t, "python.wasm", []byte("invalid module"))
	if err := Init(context.Background()); err == nil || compilationCache != nil {
		t.Fatalf("invalid module published a cache: %v", err)
	}
	installWASMCacheTestModule(t, "python.wasm", wasmTestModule(nil))
	if err := Init(context.Background()); err != nil {
		t.Fatalf("retry after initialization failure: %v", err)
	}
	if err := Close(context.Background()); err != nil || compilationCache != nil {
		t.Fatalf("cache shutdown: %v", err)
	}
	if err := Init(context.Background()); err != nil {
		t.Fatalf("restart with the existing disk cache: %v", err)
	}
}

func TestWASMWarmCacheIsolation(t *testing.T) {
	configureWASMCacheTest(t)
	config.App.WASMLimits.MaxConcurrentTasks = 8
	// Assert fresh memory and globals, mutate the global, then grow past the one-MiB limit.
	body := []byte{
		0x3f, 0, 0x41, 1, 0x47, 0x04, 0x40, 0x00, 0x0b,
		0x23, 0, 0x04, 0x40, 0x00, 0x0b,
		0x41, 1, 0x24, 0,
		0x41, 16, 0x40, 0, 0x41, 0x7f, 0x46, 0x04, 0x40, 0x00, 0x0b,
	}
	installWASMCacheTestModule(t, "python.wasm", wasmTestModule(body))
	// loop { br 0 }
	installWASMCacheTestModule(t, "javascript.wasm", wasmTestModule([]byte{0x03, 0x40, 0x0c, 0, 0x0b}))
	if err := Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			memory := int64(1 + i%2)
			task, err := NewWASMTask(context.Background(), NewTaskOptions{
				Language: "python",
				Template: &db.Tpl{
					MaxMemory: memory,
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
			output, err := task.Run(context.Background())
			if err != nil || output == nil {
				t.Errorf("task failed: %v", err)
				return
			}
			if (memory == 1 && !strings.Contains(output.Error, "unreachable")) || (memory == 2 && output.Error != "") {
				t.Errorf("shared cache changed memory limits or instance isolation: memory=%d output=%+v", memory, output)
			}
		}()
	}
	wg.Wait()

	task, err := NewWASMTask(context.Background(), NewTaskOptions{
		Language: "javascript",
		Template: &db.Tpl{
			MaxMemory: 1,
			Timeout:   1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	output, err := task.Run(context.Background())
	if err != nil || output == nil || output.Error != "execution timed out" {
		t.Fatalf("warm module lost timeout instrumentation: %+v, %v", output, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	if _, err := task.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("warm module lost request cancellation: %v", err)
	}
}

func TestWASMHostModuleCleanup(t *testing.T) {
	ctx := context.Background()
	cache := wazero.NewCompilationCache()
	defer func() { _ = cache.Close(ctx) }()
	r := &wasmRuntime{
		Runtime: wazero.NewRuntimeWithConfig(ctx, wasmRuntimeConfig(cache, 16, 0)),
	}
	builder := r.NewHostModuleBuilder("test")
	builder.NewFunctionBuilder().WithFunc(func() {}).Export("noop")
	module, err := builder.Compile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.hostModules) != 1 || r.hostModules[0] != module {
		t.Fatal("compiled host module was not tracked")
	}
	tracked := &wasmClosedModule{
		CompiledModule: module,
	}
	r.hostModules[0] = tracked

	if _, err := r.InstantiateModule(ctx, module, wazero.NewModuleConfig()); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(ctx); err != nil || len(r.hostModules) != 0 {
		t.Fatalf("host cleanup: %v", err)
	}
	if err := r.Close(ctx); err != nil {
		t.Fatalf("repeated host cleanup: %v", err)
	}
	if tracked.closes != 1 {
		t.Fatalf("compiled host module closed %d times, want 1", tracked.closes)
	}
}

type wasmClosedModule struct {
	wazero.CompiledModule
	closes int
}

func (m *wasmClosedModule) Close(ctx context.Context) error {
	m.closes++
	return m.CompiledModule.Close(ctx)
}
