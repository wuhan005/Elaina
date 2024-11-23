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
		RunCommand:    "",
	},
	{
		Name:          "python",
		Image:         "glot/python:latest",
		FileName:      "main.py",
		BuildCommands: nil,
		RunCommand:    "python main.py",
	},
	//{
	//	Name:       "go",
	//	Ext:        ".go",
	//	Image:      "elainaruntime/golang:latest",
	//	BuildCmd:   "go mod init elaina-runner && go build -v .",
	//	RunCommand: "./elaina-runner",
	//},
	//{
	//	Name:       "javascript",
	//	Ext:        ".js",
	//	Image:      "elainaruntime/javascript:latest",
	//	BuildCmd:   "",
	//	RunCommand: "node code.js",
	//},
	//{
	//	Name:       "c",
	//	Ext:        ".c",
	//	Image:      "elainaruntime/clang:latest",
	//	BuildCmd:   "gcc -v code.c -o code",
	//	RunCommand: "./code",
	//},
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
