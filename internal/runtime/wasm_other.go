//go:build !unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
)

func (*WASMTask) runWASM(context.Context, string, string, string, *wasmOutputBuffer, *wasmOutputBuffer) error {
	return errors.New("the WASI socket host requires Linux or macOS")
}
