// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"

	"github.com/wuhan005/Elaina/internal/config"
)

const wasmCompilerFeatures = api.CoreFeaturesV2 | experimental.CoreFeaturesTailCall | experimental.CoreFeaturesTypedFunctionReferences

var compilationCache wazero.CompilationCache

// Init prepares the process compilation cache before accepting requests.
func Init(ctx context.Context) error {
	if compilationCache != nil {
		return errors.New("WASM runtime is already initialized")
	}

	cache, err := wazero.NewCompilationCacheWithDir(config.App.WASMCacheDir)
	if err != nil {
		return errors.Wrap(err, "open WASM compilation cache")
	}

	if err := warmWASMModules(ctx, cache, config.App.WASMRuntimeDir); err != nil {
		_ = cache.Close(context.Background())
		return err
	}

	compilationCache = cache
	return nil
}

// Close releases the process cache after all requests have stopped.
func Close(ctx context.Context) error {
	if compilationCache == nil {
		return nil
	}

	err := compilationCache.Close(ctx)
	compilationCache = nil
	return err
}

func wasmRuntimeConfig(cache wazero.CompilationCache, memory uint32, features api.CoreFeatures) wazero.RuntimeConfig {
	// Keep cancellation instrumentation identical during warmup and execution.
	// Disable DWARF decoding, which can delay cancellation without affecting language tracebacks.
	configuration := wazero.NewRuntimeConfig().
		WithCompilationCache(cache).WithMemoryLimitPages(memory).
		WithCloseOnContextDone(true).WithDebugInfoEnabled(false)
	if features != 0 {
		configuration = configuration.WithCoreFeatures(features)
	}
	return configuration
}

func warmWASMModules(ctx context.Context, cache wazero.CompilationCache, dir string) error {
	for _, module := range []struct {
		path     string
		features api.CoreFeatures
	}{
		{
			path: "python.wasm",
		},
		{
			path: "php.wasm",
		},
		{
			path: "javascript.wasm",
		},
		{
			path: "go/compile.wasm",
		},
		{
			path: "go/link.wasm",
		},
		{
			path:     "c/clang.wasm",
			features: wasmCompilerFeatures,
		},
		{
			path:     "c/wasm-ld.wasm",
			features: wasmCompilerFeatures,
		},
	} {
		if err := ctx.Err(); err != nil {
			return err
		}

		path := filepath.Join(dir, module.path)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			logrus.WithField("module", module.path).Warn("WASM artifact is not installed; skipping warmup")
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "read WASM module %s", path)
		}

		start := time.Now()
		logrus.WithField("module", module.path).Info("Warming WASM compilation cache")
		r := wazero.NewRuntimeWithConfig(ctx, wasmRuntimeConfig(cache, 65536, module.features))
		// Retain one compiled reference until cache shutdown; never instantiate or run guest code here.
		_, err = r.CompileModule(ctx, data)
		_ = r.Close(context.Background())
		if err != nil {
			return errors.Wrapf(err, "warm WASM module %s", path)
		}
		logrus.WithField("module", module.path).WithField("duration", time.Since(start)).Info("WASM compilation cache ready")
	}
	return ctx.Err()
}
