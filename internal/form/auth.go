// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/wuhan005/Elaina/internal/auth"
)

type SignIn struct {
	Username string `json:"username" valid:"required"`
	// Password must contain 8-72 UTF-8 bytes.
	Password string `json:"password" valid:"required"`
} // @name SignInForm

type SignInTwoFactor struct {
	// A 6-digit TOTP code or an unused recovery code.
	Code string `json:"code" valid:"required"`
} // @name SignInTwoFactorForm

type ChangePassword struct {
	CurrentPassword string `json:"currentPassword" valid:"required"`
	// Password must contain 8-72 UTF-8 bytes.
	NewPassword string `json:"newPassword" valid:"required"`
} // @name ChangePasswordForm

func (f ChangePassword) Validate() error {
	return auth.ValidatePassword(f.NewPassword)
}

type ConfirmPassword struct {
	// The current password of the signed-in user.
	Password string `json:"password" valid:"required"`
} // @name ConfirmPasswordForm

type EnableTwoFactor struct {
	// A 6-digit code from the authenticator app.
	Code string `json:"code" valid:"required"`
} // @name EnableTwoFactorForm
