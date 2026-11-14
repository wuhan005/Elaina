// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"

	"github.com/wuhan005/Elaina/internal/db"
)

// ExecErrorTimeout is the ExecOutput.Error of a task that ran past its timeout.
const ExecErrorTimeout = "execution timed out"

// ExecOutput contains the program output and execution duration.
type ExecOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error"`
	Duration int64  `json:"duration"`
}

type ExecRuntime interface {
	Run(ctx context.Context) (*ExecOutput, error)
}

type NewTaskOptions struct {
	Language string
	Template *db.Tpl
	Code     []byte
	Client   string
}

// NewTask creates a WASM task for the selected language.
func NewTask(ctx context.Context, options NewTaskOptions) (ExecRuntime, error) {
	return NewWASMTask(ctx, options)
}
