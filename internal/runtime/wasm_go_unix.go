//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"encoding/pem"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tetratelabs/wazero"
)

func (t *WASMTask) runGoWASM(ctx context.Context, cache wazero.CompilationCache, toolchain, dir string, stdout, stderr *wasmOutputBuffer) error {
	imports, err := os.ReadFile(filepath.Join(toolchain, "importcfg"))
	if err != nil {
		return errors.Wrap(err, "read Go toolchain (run scripts/build-wasm-go.sh)")
	}

	file, err := parser.ParseFile(token.NewFileSet(), "main.go", t.code, parser.ImportsOnly)
	if err != nil {
		return errors.Wrap(err, "parse Go imports")
	}
	if file.Name.Name != "main" {
		return errors.New("Go source must declare package main")
	}

	http := false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return errors.Wrap(err, "parse Go import path")
		}
		if path != "unsafe" && (!strings.Contains(string(imports), "packagefile "+path+"=") || strings.Contains("/"+path+"/", "/internal/") || strings.HasPrefix(path, "vendor/")) {
			return errors.Errorf("Go package %q is not available in the WASM toolchain", path)
		}
		http = http || path == "net/http"
	}

	args := []string{"compile", "-p", "main", "-pack", "-c", "1", "-nolocalimports", "-importcfg", "/toolchain/importcfg", "-o", "/main.a", "/main.go"}
	if http {
		// Install the WASI dialer for the standard HTTP client's default transport.
		bridge := []byte("package main\nimport _ \"github.com/stealthrocket/net/http\"\n")
		if err := writeWASMWorkspaceFile(dir, "http_bridge.go", bridge); err != nil {
			return errors.Wrap(err, "write Go HTTP bridge")
		}
		args = append(args, "/http_bridge.go")
	}

	env := []string{"GOOS=wasip1", "GOARCH=wasm", "GOROOT=/toolchain", "GOMAXPROCS=1", "TMPDIR=/"}
	for _, args := range [][]string{args, {"link", "-s", "-w", "-linkmode=internal", "-importcfg", "/toolchain/importcfg", "-o", "/main.wasm", "/main.a"}} {
		if err := t.runWASMModule(ctx, cache, wasmModuleOptions{
			path:      filepath.Join(toolchain, args[0]+".wasm"),
			args:      args,
			env:       env,
			dir:       dir,
			toolchain: toolchain,
		}, stderr, stderr); err != nil {
			return errors.Wrapf(err, "Go %s", args[0])
		}
	}

	var runtimeEnv []string
	if t.template.InternetAccess {
		certificates, err := wasmGoRootCertificates()
		if err != nil {
			return err
		}
		if len(certificates) > 0 {
			if err := writeWASMWorkspaceFile(dir, ".elaina-ca.pem", certificates); err != nil {
				return errors.Wrap(err, "write Go root certificates")
			}
			runtimeEnv = []string{"SSL_CERT_FILE=/.elaina-ca.pem"}
		}
	}

	// Request-specific programs must not accumulate in the persistent toolchain cache.
	return t.runWASMModule(ctx, nil, wasmModuleOptions{
		path:     filepath.Join(dir, "main.wasm"),
		args:     []string{"main"},
		dir:      dir,
		internet: t.template.InternetAccess,
		env:      runtimeEnv,
	}, stdout, stderr)
}

func wasmGoRootCertificates() ([]byte, error) {
	for _, path := range []string{"/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/cert.pem", "/etc/pki/tls/certs/ca-bundle.crt"} {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, errors.Wrap(err, "read Go root certificates")
		}

		var certificates []byte
		for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
			// Only public certificates are exposed to guest code.
			if block.Type == "CERTIFICATE" {
				certificates = append(certificates, pem.EncodeToMemory(block)...)
			}
		}

		if len(certificates) > 0 {
			return certificates, nil
		}
	}

	return nil, nil
}
