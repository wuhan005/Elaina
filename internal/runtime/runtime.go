// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
)

type ExecInputRunInstructions struct {
	BuildCommands []string `json:"buildCommands"`
	RunCommand    string   `json:"runCommand"`
}

type ExecInputFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ExecInput struct {
	RunInstructions ExecInputRunInstructions `json:"runInstructions"`
	Files           []*ExecInputFile         `json:"files"`
	Stdin           *string                  `json:"stdin"`
}

// ExecOutput is the glot output style.
type ExecOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error"`
	Duration int64  `json:"duration"`
}

type ExecRuntime interface {
	Run(ctx context.Context) (*ExecOutput, error)
}
