// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dto

import (
	"time"

	"github.com/wuhan005/Elaina/internal/db"
)

type User struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Username               string     `json:"username"`
	TwoFactorEnabled       bool       `json:"twoFactorEnabled"`
	RecoveryCodesRemaining int        `json:"recoveryCodesRemaining"`
	LastSignInAt           *time.Time `json:"lastSignInAt" extensions:"x-nullable"`
} // @name User

func ToUser(user *db.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:                     user.ID,
		CreatedAt:              user.CreatedAt,
		UpdatedAt:              user.UpdatedAt,
		Username:               user.Username,
		TwoFactorEnabled:       user.TOTPEnabled,
		RecoveryCodesRemaining: len(user.RecoveryCodes),
		LastSignInAt:           user.LastSignInAt,
	}
}

func ToUsers(users []*db.User) []*User {
	items := make([]*User, 0, len(users))
	for _, user := range users {
		items = append(items, ToUser(user))
	}
	return items
}

type SignInResp struct {
	// Session token for the Authorization header ("Token <token>").
	Token string `json:"token"`
	// When true, the token can only be used to submit a verification code to signInTwoFactor.
	TwoFactorRequired bool `json:"twoFactorRequired"`
} // @name SignInResp

type TwoFactorSetupResp struct {
	// Base32 TOTP secret for manual entry.
	Secret string `json:"secret"`
	// otpauth:// URL encoded in the QR code.
	URL string `json:"url"`
	// PNG QR code as a data URL.
	QRCode string `json:"qrCode"`
} // @name TwoFactorSetupResp

type RecoveryCodesResp struct {
	// One-time recovery codes. They are only shown once.
	Codes []string `json:"codes"`
} // @name RecoveryCodesResp
