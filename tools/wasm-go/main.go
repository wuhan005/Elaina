// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if err := build(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func build() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run . <output directory>")
	}
	dir, err := filepath.Abs(os.Args[1])
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Native compiler processes run only while preparing the deployment artifacts.
	run := func(args ...string) ([]byte, error) {
		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0", "GOWORK=off", "GOFLAGS=", "GOEXPERIMENT=", "GOTOOLCHAIN=go1.27.3")
		cmd.Stderr = os.Stderr
		return cmd.Output()
	}
	for _, tool := range []string{"compile", "link"} {
		if _, err := run("build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(dir, tool+".wasm"), "cmd/"+tool); err != nil {
			return fmt.Errorf("build %s: %w", tool, err)
		}
	}
	data, err := run("list", "-trimpath", "-tags=getaddrinfo", "-export", "-deps", "-json", "std", "github.com/stealthrocket/net/http")
	if err != nil {
		return fmt.Errorf("build Go libraries: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var imports strings.Builder
	licenses := map[string]string{"Go-LICENSE": filepath.Join(runtime.GOROOT(), "LICENSE")}
	for {
		var pkg struct {
			ImportPath, Export string
			Module             *struct{ Path, Dir string }
		}
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if pkg.Export == "" {
			continue
		}
		if pkg.Module != nil && pkg.Module.Path == "github.com/stealthrocket/net" {
			for _, name := range []string{"LICENSE", "NOTICE"} {
				licenses["stealthrocket-net-"+name] = filepath.Join(pkg.Module.Dir, name)
			}
		}
		path := filepath.Join(dir, "pkg", filepath.FromSlash(pkg.ImportPath)+".a")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		archive, err := os.ReadFile(pkg.Export)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, archive, 0644); err != nil {
			return err
		}
		fmt.Fprintf(&imports, "packagefile %s=/toolchain/pkg/%s.a\n", pkg.ImportPath, pkg.ImportPath)
	}
	if err := os.WriteFile(filepath.Join(dir, "importcfg"), []byte(imports.String()), 0644); err != nil {
		return err
	}
	for name, path := range licenses {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			return err
		}
	}
	version, err := run("version")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "VERSION"), version, 0644)
}
