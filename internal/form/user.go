// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/wuhan005/Elaina/internal/auth"
)

type CreateUser struct {
	// 2-32 letters, digits, '_', '.' or '-', starting with a letter or digit.
	Username string `json:"username" valid:"required"`
	// Password must contain 8-72 UTF-8 bytes.
	Password string `json:"password" valid:"required"`
} // @name CreateUserReq

func (f CreateUser) Validate() error {
	if err := auth.ValidateUsername(f.Username); err != nil {
		return err
	}
	return auth.ValidatePassword(f.Password)
}

type ResetPassword struct {
	// Password must contain 8-72 UTF-8 bytes.
	Password string `json:"password" valid:"required"`
} // @name ResetPasswordReq

func (f ResetPassword) Validate() error {
	return auth.ValidatePassword(f.Password)
}
