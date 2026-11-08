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
	_ "github.com/wuhan005/Elaina/internal/dto" // Swagger response schemas.
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

// SignIn sign in with the administrator password.
// @Summary Sign in with the administrator password
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body form.SignIn true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 403 {object} dto.ErrorResponse "Invalid password"
// @ID signIn
// @Router /api/auth/sign-in [post]
func (h *AuthHandler) SignIn(ctx context.Context, session session.Session, f form.SignIn) error {
	if len(f.Password) < 8 || len(f.Password) > 72 ||
		bcrypt.CompareHashAndPassword([]byte(config.App.PasswordHash), []byte(f.Password)) != nil {
		return ctx.Error(http.StatusForbidden, "Invalid password")
	}

	session.Set(context.SessionIDIsAuthenticated, true)

	return ctx.Success(session.ID())
}

// Profile check the current session.
// @Summary Check the current session
// @Tags Auth
// @Description Returns data: null when the session is authenticated.
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[any]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @ID profile
// @Router /api/auth/profile [get]
func (h *AuthHandler) Profile(ctx context.Context) error {
	return ctx.Success()
}

// SignOut sign out the current session.
// @Summary Sign out the current session
// @Tags Auth
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[string]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @ID signOut
// @Router /api/auth/sign-out [post]
func (h *AuthHandler) SignOut(ctx context.Context, session session.Session) error {
	session.Flush()
	return ctx.Success("Sign out successfully")
}
