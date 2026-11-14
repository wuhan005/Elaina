// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/flamego/session"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/auth"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/form"
)

type AccountHandler struct{}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}

// ChangePassword change the password of the signed-in user.
// @Summary Change the password of the signed-in user
// @Tags Account
// @Description Other sessions of the user are signed out; the current session stays valid.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.ChangePassword true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Current password is incorrect"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID changePassword
// @Router /api/account/password [put]
func (h *AccountHandler) ChangePassword(ctx context.Context, session session.Session, f form.ChangePassword) error {
	if !auth.CheckPassword(ctx.User.PasswordHash, f.CurrentPassword) {
		return ctx.Error(http.StatusForbidden, "Current password is incorrect")
	}

	hash, err := auth.HashPassword(f.NewPassword)
	if err != nil {
		return ctx.Error(http.StatusBadRequest, "%s", err.Error())
	}
	version, err := db.Users.UpdatePassword(ctx.Request().Context(), ctx.User.ID, hash)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to update password")
		return ctx.ServerError()
	}
	session.Set(context.SessionKeySessionVersion, version)
	return ctx.Success("Password changed")
}

// SetupTwoFactor start two-factor authentication setup.
// @Summary Start two-factor authentication setup
// @Tags Account
// @Description Generates a new TOTP secret. Two-factor authentication stays disabled until enableTwoFactor confirms a code.
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[dto.TwoFactorSetupResp]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 409 {object} dto.ErrorResponse "Already enabled"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID setupTwoFactor
// @Router /api/account/2fa/setup [post]
func (h *AccountHandler) SetupTwoFactor(ctx context.Context) error {
	if ctx.User.TOTPEnabled {
		return ctx.Error(http.StatusConflict, "Two-factor authentication is already enabled")
	}

	key, err := auth.GenerateTOTP(ctx.User.Username)
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to generate TOTP key")
		return ctx.ServerError()
	}
	if err := db.Users.SetTOTPSecret(ctx.Request().Context(), ctx.User.ID, key.Secret); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to save TOTP secret")
		return ctx.ServerError()
	}
	return ctx.Success(dto.TwoFactorSetupResp{
		Secret: key.Secret,
		URL:    key.URL,
		QRCode: key.QRCode,
	})
}

// EnableTwoFactor enable two-factor authentication.
// @Summary Enable two-factor authentication
// @Tags Account
// @Description Confirms the secret from setupTwoFactor and returns new recovery codes.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.EnableTwoFactor true "Request body"
// @Success 200 {object} dto.Response[dto.RecoveryCodesResp]
// @Failure 400 {object} dto.ErrorResponse "Setup not started"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Invalid verification code"
// @Failure 409 {object} dto.ErrorResponse "Already enabled"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID enableTwoFactor
// @Router /api/account/2fa/enable [post]
func (h *AccountHandler) EnableTwoFactor(ctx context.Context, f form.EnableTwoFactor) error {
	if ctx.User.TOTPEnabled {
		return ctx.Error(http.StatusConflict, "Two-factor authentication is already enabled")
	}
	if ctx.User.TOTPSecret == "" {
		return ctx.Error(http.StatusBadRequest, "Start two-factor setup first")
	}
	if !auth.ValidateTOTP(f.Code, ctx.User.TOTPSecret) {
		return ctx.Error(http.StatusForbidden, "Invalid verification code")
	}

	codes, hashes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to generate recovery codes")
		return ctx.ServerError()
	}
	if err := db.Users.EnableTOTP(ctx.Request().Context(), ctx.User.ID, hashes); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to enable TOTP")
		return ctx.ServerError()
	}
	return ctx.Success(dto.RecoveryCodesResp{Codes: codes})
}

// DisableTwoFactor disable two-factor authentication.
// @Summary Disable two-factor authentication
// @Tags Account
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.ConfirmPassword true "Request body"
// @Success 200 {object} dto.Response[string]
// @Failure 400 {object} dto.ErrorResponse "Not enabled"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Password is incorrect"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID disableTwoFactor
// @Router /api/account/2fa/disable [post]
func (h *AccountHandler) DisableTwoFactor(ctx context.Context, f form.ConfirmPassword) error {
	if !ctx.User.TOTPEnabled {
		return ctx.Error(http.StatusBadRequest, "Two-factor authentication is not enabled")
	}
	if !auth.CheckPassword(ctx.User.PasswordHash, f.Password) {
		return ctx.Error(http.StatusForbidden, "Password is incorrect")
	}
	if err := db.Users.DisableTOTP(ctx.Request().Context(), ctx.User.ID, false); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to disable TOTP")
		return ctx.ServerError()
	}
	return ctx.Success("Two-factor authentication disabled")
}

// RegenerateRecoveryCodes regenerate recovery codes.
// @Summary Regenerate recovery codes
// @Tags Account
// @Description Replaces all existing recovery codes.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.ConfirmPassword true "Request body"
// @Success 200 {object} dto.Response[dto.RecoveryCodesResp]
// @Failure 400 {object} dto.ErrorResponse "Not enabled"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Password is incorrect"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID regenerateRecoveryCodes
// @Router /api/account/2fa/recovery-codes [post]
func (h *AccountHandler) RegenerateRecoveryCodes(ctx context.Context, f form.ConfirmPassword) error {
	if !ctx.User.TOTPEnabled {
		return ctx.Error(http.StatusBadRequest, "Two-factor authentication is not enabled")
	}
	if !auth.CheckPassword(ctx.User.PasswordHash, f.Password) {
		return ctx.Error(http.StatusForbidden, "Password is incorrect")
	}

	codes, hashes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to generate recovery codes")
		return ctx.ServerError()
	}
	if err := db.Users.SetRecoveryCodes(ctx.Request().Context(), ctx.User.ID, hashes); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to save recovery codes")
		return ctx.ServerError()
	}
	return ctx.Success(dto.RecoveryCodesResp{Codes: codes})
}
