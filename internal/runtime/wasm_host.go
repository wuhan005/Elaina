// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"

	"github.com/tetratelabs/wazero"
)

// wasi-go does not close its compiled host modules; a shared cache would retain them after each task.
type wasmRuntime struct {
	wazero.Runtime
	hostModules []wazero.CompiledModule
}

func (r *wasmRuntime) NewHostModuleBuilder(name string) wazero.HostModuleBuilder {
	return &wasmHostModuleBuilder{
		HostModuleBuilder: r.Runtime.NewHostModuleBuilder(name),
		runtime:           r,
	}
}

func (r *wasmRuntime) Close(ctx context.Context) error {
	err := r.Runtime.Close(ctx)
	for _, module := range r.hostModules {
		_ = module.Close(ctx)
	}
	r.hostModules = nil
	return err
}

type wasmHostModuleBuilder struct {
	wazero.HostModuleBuilder
	runtime *wasmRuntime
}

func (b *wasmHostModuleBuilder) Compile(ctx context.Context) (wazero.CompiledModule, error) {
	module, err := b.HostModuleBuilder.Compile(ctx)
	if err == nil {
		b.runtime.hostModules = append(b.runtime.hostModules, module)
	}
	return module, err
}
