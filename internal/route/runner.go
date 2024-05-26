// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"context"
)

type RunnerHandler struct{}

func NewRunnerHandler() *RunnerHandler {
	return &RunnerHandler{}
}

func (h *RunnerHandler) Runner(ctx context.Context) {

}

func (h *RunnerHandler) View(ctx context.Context) {

}

func (h *RunnerHandler) Execute(ctx context.Context) {

}
