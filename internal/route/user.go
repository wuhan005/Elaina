// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/auth"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/form"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Userer loads the user from the id path parameter and rejects managing the signed-in user.
func (h *UserHandler) Userer(ctx context.Context) error {
	userID := uint(ctx.ParamInt("id"))
	user, err := db.Users.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return ctx.Error(http.StatusNotFound, "user::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get user")
		return ctx.ServerError()
	}
	if user.ID == ctx.User.ID {
		return ctx.Error(http.StatusBadRequest, "user::manage_self")
	}

	ctx.Map(user)
	return nil
}

// List list users.
// @Summary List users
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[[]dto.User]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID listUsers
// @Router /api/users [get]
func (h *UserHandler) List(ctx context.Context) error {
	users, err := db.Users.List(ctx.Request().Context())
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to list users")
		return ctx.ServerError()
	}
	return ctx.Success(dto.ToUsers(users))
}

// Create create a user.
// @Summary Create a user
// @Tags Users
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.CreateUser true "Request body"
// @Success 200 {object} dto.Response[dto.User]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 409 {object} dto.ErrorResponse "Username already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID createUser
// @Router /api/users [post]
func (h *UserHandler) Create(ctx context.Context, f form.CreateUser) error {
	hash, err := auth.HashPassword(f.Password)
	if err != nil {
		return ctx.ErrorFrom(http.StatusBadRequest, err)
	}
	user, err := db.Users.Create(ctx.Request().Context(), db.CreateUserOptions{
		Username:     f.Username,
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, db.ErrUserExists) {
			return ctx.Error(http.StatusConflict, "user::exists")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create user")
		return ctx.ServerError()
	}
	return ctx.Success(dto.ToUser(user))
}

// Delete delete a user.
// @Summary Delete a user
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Cannot delete yourself"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID deleteUser
// @Router /api/user/{id} [delete]
func (h *UserHandler) Delete(ctx context.Context, user *db.User) error {
	if err := db.Users.Delete(ctx.Request().Context(), user.ID); err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return ctx.Error(http.StatusNotFound, "user::not_found")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to delete user")
		return ctx.ServerError()
	}
	return ctx.Success(ctx.Tr("user::deleted"))
}

// ResetPassword set a new password for a user.
// @Summary Set a new password for a user
// @Tags Users
// @Description Signs the user out of all sessions.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Param data body form.ResetPassword true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID resetUserPassword
// @Router /api/user/{id}/password [put]
func (h *UserHandler) ResetPassword(ctx context.Context, user *db.User, f form.ResetPassword) error {
	hash, err := auth.HashPassword(f.Password)
	if err != nil {
		return ctx.ErrorFrom(http.StatusBadRequest, err)
	}
	if _, err := db.Users.UpdatePassword(ctx.Request().Context(), user.ID, hash); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to reset password")
		return ctx.ServerError()
	}
	return ctx.Success(ctx.Tr("user::password_updated"))
}

// ResetTwoFactor turn off two-factor authentication for a user.
// @Summary Turn off two-factor authentication for a user
// @Tags Users
// @Description For users who lost their authenticator and recovery codes. Signs the user out of all sessions.
// @Produce json
// @Security SessionAuth
// @Param id path int true "ID"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID resetUserTwoFactor
// @Router /api/user/{id}/2fa [delete]
func (h *UserHandler) ResetTwoFactor(ctx context.Context, user *db.User) error {
	if err := db.Users.DisableTOTP(ctx.Request().Context(), user.ID, true); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to reset TOTP")
		return ctx.ServerError()
	}
	return ctx.Success(ctx.Tr("user::two_factor_reset"))
}
