//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	preview1 "github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wazergo"
	"github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type wasmModuleOptions struct {
	path      string
	args      []string
	env       []string
	dir       string
	toolchain string
	internet  bool
	features  api.CoreFeatures
}

func (t *WASMTask) runWASM(ctx context.Context, runtimeDir, cacheDir, dir string, stdout, stderr *wasmOutputBuffer) error {
	cache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		return errors.Wrap(err, "open WASM compilation cache")
	}
	defer func() { _ = cache.Close(ctx) }()

	if t.runner.Name == "go" {
		return t.runGoWASM(ctx, cache, filepath.Join(runtimeDir, "go"), dir, stdout, stderr)
	}

	if t.runner.Name == "c" {
		toolchain := filepath.Join(runtimeDir, "c")
		if err := t.compileCWASM(ctx, cache, toolchain, dir, stderr); err != nil {
			return err
		}

		return t.runWASMModule(ctx, nil, wasmModuleOptions{
			path: filepath.Join(dir, "main.wasm"), args: []string{"main"}, dir: dir, internet: t.template.InternetAccess,
		}, stdout, stderr)
	}

	return t.runWASMModule(ctx, cache, wasmModuleOptions{
		path: filepath.Join(runtimeDir, t.runner.WASMModule), args: []string{t.runner.Name, "/" + t.runner.FileName},
		dir: dir, internet: t.template.InternetAccess,
	}, stdout, stderr)
}

func (t *WASMTask) runWASMModule(ctx context.Context, cache wazero.CompilationCache, options wasmModuleOptions, stdout, stderr *wasmOutputBuffer) error {
	memory := t.template.MaxMemory
	if memory == 0 {
		memory = 4096
	}

	// Interpreter DWARF decoding can delay cancellation; language tracebacks do not depend on it.
	configuration := wazero.NewRuntimeConfig().
		WithCompilationCache(cache).WithMemoryLimitPages(uint32(memory * 16)).
		WithCloseOnContextDone(true).WithDebugInfoEnabled(false)
	if options.features != 0 {
		configuration = configuration.WithCoreFeatures(options.features)
	}
	r := wazero.NewRuntimeWithConfig(ctx, configuration)
	defer func() { _ = r.Close(ctx) }()

	data, err := os.ReadFile(options.path)
	if err != nil {
		return errors.Wrap(err, "read WASM module (install the language artifacts first)")
	}
	module, err := r.CompileModule(ctx, data)
	if err != nil {
		return errors.Wrap(err, "compile WASM module")
	}
	defer func() { _ = module.Close(ctx) }()

	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return errors.Wrap(err, "open WASM stdio")
	}
	defer func() { _ = null.Close() }()

	var host *wasmunix.System
	dirs := []string{options.dir}
	preopens := map[wasi.FD]string{3: "/"}
	if options.toolchain != "" {
		dirs = append(dirs, options.toolchain+":ro")
		preopens[4] = "/toolchain"
	}

	ctx, system, err := imports.NewBuilder().WithName(options.args[0]).
		WithArgs(options.args[1:]...).WithEnv(append([]string{"PWD=/"}, options.env...)...).WithDirs(dirs...).
		WithStdio(int(null.Fd()), int(null.Fd()), int(null.Fd())).
		WithMaxOpenFiles(256).WithMaxOpenDirs(64).WithSocketsExtension("auto", module).
		WithDecorators(wazergo.DecoratorFunc(wasmSocketABI)).
		WithWrappers(func(system wasi.System) wasi.System {
			host = system.(*wasmunix.System)
			return &wasmSystem{System: &wasmNetworkSystem{System: host}, internet: options.internet, preopens: preopens,
				outputs: map[wasi.FD]*wasmOutputBuffer{1: stdout, 2: stderr}}
		}).Instantiate(ctx, r)
	if err != nil {
		return errors.Wrap(err, "instantiate WASI host")
	}
	defer func() { _ = system.Close(ctx) }()

	if options.toolchain != "" {
		if errno := host.FDStatSetRights(ctx, 4, wasmReadOnlyRights&wasi.DirectoryRights, wasmReadOnlyRights); errno != wasi.ESUCCESS {
			return errors.Errorf("restrict WASM toolchain: %s", errno)
		}
	}

	// Wake WASI timers and polls; socket calls use Go's cancellable network poller.
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		_ = host.Shutdown(ctx)
		close(done)
	})
	defer func() {
		if !stop() {
			<-done
		}
	}()

	_, err = r.InstantiateModule(ctx, module, wazero.NewModuleConfig().WithName(""))
	return err
}

func wasmSocketABI(_ string, fn wazergo.Function[*preview1.Module]) wazergo.Function[*preview1.Module] {
	if fn.Name == "sock_getaddrinfo" {
		return fn.WithFunc(func(m *preview1.Module, ctx context.Context, module api.Module, stack []uint64) {
			// wasi-go allocates host memory from this guest-controlled result count.
			if uint32(stack[6]) > 256 {
				stack[0] = uint64(wasi.EINVAL)
				return
			}

			fn.Func(m, ctx, module, stack)
		})
	}

	if fn.Name == "sock_recv" {
		// wasi-go v0.8.0 writes four bytes for WASI's two-byte roflags value.
		fixed := wazergo.F5(wasmSocketReceive)
		fixed.Name = fn.Name
		return fixed
	}

	return fn
}

func wasmSocketReceive(m *preview1.Module, ctx context.Context, fd types.Int32, buffers types.List[wasi.IOVec], flags types.Int32, count types.Pointer[types.Int32], resultFlags types.Pointer[types.Uint16]) types.Errno {
	n, receivedFlags, errno := m.WASI.SockRecv(ctx, wasi.FD(fd), buffers.Slice(), wasi.RIFlags(flags))
	if errno == wasi.ESUCCESS {
		count.Store(types.Int32(n))
		resultFlags.Store(types.Uint16(receivedFlags))
	}

	return types.Errno(errno)
}
