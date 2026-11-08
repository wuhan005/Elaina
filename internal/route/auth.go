// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/flamego/session"
	"golang.org/x/crypto/bcrypt"

	"github.com/wuhan005/Elaina/internal/config"
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
	if len(f.Password) < 8 || len(f.Password) > 72 ||
		bcrypt.CompareHashAndPassword([]byte(config.App.PasswordHash), []byte(f.Password)) != nil {
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
