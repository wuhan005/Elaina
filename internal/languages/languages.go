// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package languages

var runners = []Runner{
	{
		Name:          "php",
		Image:         "glot/php:latest",
		FileName:      "main.php",
		BuildCommands: nil,
		RunCommand:    "php main.php",
	},
	{
		Name:          "python",
		Image:         "glot/python:latest",
		FileName:      "main.py",
		BuildCommands: nil,
		RunCommand:    "python main.py",
	},
	{
		Name:          "go",
		Image:         "glot/go:latest",
		FileName:      "main.go",
		BuildCommands: nil,
		RunCommand:    "go run main.go",
	},
	{
		Name:          "javascript",
		Image:         "glot/javascript:latest",
		FileName:      "main.js",
		BuildCommands: nil,
		RunCommand:    "node main.js",
	},
	{
		Name:          "javascript",
		Image:         "glot/c:latest",
		FileName:      "main.c",
		BuildCommands: nil,
		RunCommand:    "clang main.c && ./a.out",
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
