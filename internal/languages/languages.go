// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package languages

type Runner struct {
	Name string
	// Label is the display name shown on the runner page.
	Label      string
	FileName   string
	WASMModule string
}

var runners = []Runner{
	{
		Name:       "php",
		Label:      "PHP",
		FileName:   "main.php",
		WASMModule: "php.wasm",
	},
	{
		Name:       "python",
		Label:      "Python",
		FileName:   "main.py",
		WASMModule: "python.wasm",
	},
	{
		Name:       "go",
		Label:      "Go",
		FileName:   "main.go",
		WASMModule: "go/compile.wasm",
	},
	{
		Name:       "javascript",
		Label:      "JavaScript",
		FileName:   "main.js",
		WASMModule: "javascript.wasm",
	},
	{
		Name:       "c",
		Label:      "C",
		FileName:   "main.c",
		WASMModule: "c/clang.wasm",
	},
}

func Get(name string) (*Runner, bool) {
	for _, r := range runners {
		if r.Name == name {
			return &r, true
		}
	}
	return nil, false
}
