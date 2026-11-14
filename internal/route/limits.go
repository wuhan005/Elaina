// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/flamego/flamego"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/dto"
)

func limitRequestBody(ctx flamego.Context) {
	r := ctx.Request().Request
	if r.Body == nil {
		return
	}

	limit := config.App.WASMLimits.MaxRequestBytes
	var data []byte
	var err error
	if r.ContentLength > limit {
		err = &http.MaxBytesError{Limit: limit}
	} else {
		// Read before decoding so chunked bodies and trailing JSON cannot bypass the limit.
		data, err = io.ReadAll(http.MaxBytesReader(ctx.ResponseWriter(), r.Body, limit))
	}
	_ = r.Body.Close()
	if err != nil {
		status, message := http.StatusBadRequest, "Failed to read request body"
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			status, message = http.StatusRequestEntityTooLarge, "Request body exceeds the configured size limit"
		}
		context.WriteJSON(ctx.ResponseWriter(), r, status, dto.ErrorResponse{
			Msg: message,
		})
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
}
