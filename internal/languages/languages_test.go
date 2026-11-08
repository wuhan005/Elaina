// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package languages

import "testing"

func TestRunners(t *testing.T) {
	names := make(map[string]bool)
	for _, runner := range runners {
		if names[runner.Name] {
			t.Fatalf("duplicate language: %s", runner.Name)
		}
		names[runner.Name] = true
	}
	for _, name := range []string{"php", "python", "javascript", "go", "c"} {
		runner, ok := Get(name)
		if !ok {
			t.Fatalf("missing language: %s", name)
		}
		if runner.WASMModule == "" {
			t.Fatalf("unexpected WASM configuration for %s", name)
		}
	}
}
