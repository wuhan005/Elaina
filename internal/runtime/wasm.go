// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tetratelabs/wazero/sys"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/languages"
)

const wasmOutputLimit = 1024 * 1024

var ErrSourceLimit = errors.New("source code exceeds the configured size limit")

var _ ExecRuntime = (*WASMTask)(nil)

type WASMTask struct {
	runner   *languages.Runner
	template *db.Tpl
	code     []byte
	client   string
}

// NewWASMTask creates a wazero task inside the Go backend.
func NewWASMTask(_ context.Context, options NewTaskOptions) (*WASMTask, error) {
	runner, ok := languages.Get(options.Language)
	if !ok || runner.WASMModule == "" {
		return nil, errors.Errorf("unsupported WASM language: %q", options.Language)
	}

	if options.Template == nil {
		return nil, errors.New("missing template")
	}

	if int64(len(options.Code)) > config.App.WASMLimits.MaxSourceBytes {
		return nil, ErrSourceLimit
	}

	if options.Template.MaxMemory < 0 || options.Template.MaxMemory > 4096 {
		return nil, errors.New("WASM memory limit must be between 0 and 4096 MiB")
	}

	if options.Template.Timeout < 0 || options.Template.Timeout > 60 {
		return nil, errors.New("WASM timeout must be between 0 and 60 seconds")
	}

	return &WASMTask{
		runner:   runner,
		template: options.Template,
		code:     options.Code,
		client:   options.Client,
	}, nil
}

func (t *WASMTask) Run(ctx context.Context) (*ExecOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if int64(len(t.code)) > config.App.WASMLimits.MaxSourceBytes {
		return nil, ErrSourceLimit
	}

	release, err := taskAdmission.acquire(ctx, t.client)
	if err != nil {
		return nil, err
	}
	defer release()

	timeout := t.template.Timeout
	if timeout == 0 {
		timeout = 10
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	runtimeDir, err := filepath.Abs(config.App.WASMRuntimeDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolve WASM runtime directory")
	}

	dir, err := os.MkdirTemp("", "elaina-wasm-")
	if err != nil {
		return nil, errors.Wrap(err, "create WASM workspace")
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("Failed to remove WASM workspace")
		}
	}()

	if err := os.WriteFile(filepath.Join(dir, t.runner.FileName), t.code, 0600); err != nil {
		return nil, errors.Wrap(err, "write WASM source")
	}

	stdout := &wasmOutputBuffer{
		limit:  wasmOutputLimit,
		cancel: cancel,
	}
	stderr := &wasmOutputBuffer{
		limit:  wasmOutputLimit,
		cancel: cancel,
	}

	startAt := time.Now()
	runErr := t.runWASM(execCtx, runtimeDir, dir, stdout, stderr)
	output := &ExecOutput{
		Stdout:   stdout.buffer.String(),
		Stderr:   stderr.buffer.String(),
		Duration: time.Since(startAt).Nanoseconds(),
	}

	var exitErr *sys.ExitError
	switch {
	case ctx.Err() != nil:
		return nil, errors.Wrap(ctx.Err(), "run WASM task")
	case stdout.exceeded || stderr.exceeded:
		output.Error = "output limit exceeded"
	case execCtx.Err() != nil:
		output.Error = ExecErrorTimeout
	case errors.As(runErr, &exitErr):
		if exitErr.ExitCode() != 0 {
			output.Error = fmt.Sprintf("exit status %d", exitErr.ExitCode())
		}
	case runErr != nil:
		output.Error = runErr.Error()
	}

	return output, nil
}

type wasmOutputBuffer struct {
	buffer   bytes.Buffer
	limit    int
	cancel   context.CancelFunc
	exceeded bool
}

func (b *wasmOutputBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - b.buffer.Len()
	if len(p) <= remaining {
		return b.buffer.Write(p)
	}

	n, _ := b.buffer.Write(p[:remaining])
	b.exceeded = true
	b.cancel()
	return n, errors.New("output limit exceeded")
}
