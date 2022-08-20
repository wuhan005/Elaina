// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"context"
)

// CommandOutput contains the body and the exit code of the command execution.
type CommandOutput struct {
	ExitCode int    `json:"exit_code"`
	Body     []byte `json:"body"`
}

type Runner interface {
	Run(ctx context.Context) ([]*CommandOutput, error)
}
