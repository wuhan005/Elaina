// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/flamego/session"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/form"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Authenticator(ctx context.Context) error {
	if !ctx.IsAuthenticated {
		return ctx.Error(http.StatusUnauthorized, "Unauthorized")
	}
	return nil
}

func (h *AuthHandler) SignIn(ctx context.Context, session session.Session, f form.SignIn) error {
	appPassword := os.Getenv("APP_PASSWORD")
	if subtle.ConstantTimeCompare([]byte(appPassword), []byte(f.Password)) != 1 {
		return ctx.Error(http.StatusForbidden, "Invalid password")
	}

	session.Set(context.SessionIDIsAuthenticated, true)

	return ctx.Success(session.ID())
}

func (h *AuthHandler) Profile(ctx context.Context) error {
	return ctx.Success()
}

func (h *AuthHandler) SignOut(ctx context.Context, session session.Session) error {
	session.Flush()
	return ctx.Success("Sign out successfully")
}
