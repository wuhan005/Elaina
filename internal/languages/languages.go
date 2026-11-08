// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package languages

type Runner struct {
	Name       string
	FileName   string
	WASMModule string
}

var runners = []Runner{
	{
		Name:       "php",
		FileName:   "main.php",
		WASMModule: "php.wasm",
	},
	{
		Name:       "python",
		FileName:   "main.py",
		WASMModule: "python.wasm",
	},
	{
		Name:       "go",
		FileName:   "main.go",
		WASMModule: "go/compile.wasm",
	},
	{
		Name:       "javascript",
		FileName:   "main.js",
		WASMModule: "javascript.wasm",
	},
	{
		Name:       "c",
		FileName:   "main.c",
		WASMModule: "c/clang.wasm",
	},
}

func Get(name string) (*Runner, bool) {
	for _, r := range runners {
		r := r
		if r.Name == name {
			return &r, true
		}
	}

	return nil, false
}
