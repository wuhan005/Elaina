// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/wuhan005/Elaina/internal/context"
)

type BaseHandler struct{}

func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

func (h *BaseHandler) Index(ctx context.Context) string {
	return ""
}
