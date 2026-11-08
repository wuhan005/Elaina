// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

type Response[T any] struct {
	Data T `json:"data"`
}

type ErrorResponse struct {
	Msg string `json:"msg"`
} // @name ErrorResponse
