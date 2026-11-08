//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

func (t *WASMTask) compileCWASM(ctx context.Context, cache wazero.CompilationCache, toolchain, dir string, stderr *wasmOutputBuffer) error {
	var commands [][]string
	for _, source := range []struct{ path, output string }{{"/main.c", "/main.o"}, {"/toolchain/network.c", "/network.o"}} {
		commands = append(commands, []string{"clang", "-cc1", "-triple", "wasm32-unknown-wasi", "-isysroot", "/toolchain/sysroot",
			"-I/toolchain/include",
			"-internal-isystem", "/toolchain/sysroot/include/wasm32-wasip1", "-internal-isystem", "/toolchain/sysroot/lib/clang/22/include",
			"-std=c17", "-O0", "-ferror-limit", "10", "-emit-obj", "-o", source.output, "-x", "c", source.path})
	}
	commands = append(commands, []string{"wasm-ld", "-z", "stack-size=1048576", "--strip-debug", "-L/toolchain/sysroot/lib/wasm32-wasip1",
		"/toolchain/sysroot/lib/wasm32-wasip1/crt1.o", "/main.o", "/network.o", "-lc", "-lclang_rt.builtins", "-o", "/main.wasm"})
	for _, args := range commands {
		// Compiler stages never inherit the program's network permission or environment.
		if err := t.runWASMModule(ctx, cache, wasmModuleOptions{
			path: filepath.Join(toolchain, args[0]+".wasm"), args: args, env: []string{"TMPDIR=/"}, dir: dir, toolchain: toolchain,
			// These instructions are enabled only for the pinned compiler and linker.
			features: api.CoreFeaturesV2 | experimental.CoreFeaturesTailCall | experimental.CoreFeaturesTypedFunctionReferences,
		}, stderr, stderr); err != nil {
			return errors.Wrapf(err, "C %s", args[0])
		}
	}
	return nil
}
