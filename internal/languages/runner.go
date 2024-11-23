// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package languages

type Runner struct {
	Name          string
	Image         string
	FileName      string
	BuildCommands []string
	RunCommand    string
}
