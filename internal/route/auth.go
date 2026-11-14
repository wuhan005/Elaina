// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"time"

	"github.com/flamego/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/auth"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/form"
)

const (
	twoFactorTimeout     = 5 * time.Minute
	maxTwoFactorAttempts = 5
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Authenticator requires a signed-in session and loads the current user. The session is
// flushed when the user was deleted or the session version is outdated.
func (h *AuthHandler) Authenticator(ctx context.Context, session session.Session) error {
	if !ctx.IsAuthenticated {
		return ctx.Error(http.StatusUnauthorized, "Unauthorized")
	}

	user, err := db.Users.GetByID(ctx.Request().Context(), ctx.UserID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			session.Flush()
			return ctx.Error(http.StatusUnauthorized, "Unauthorized")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get user")
		return ctx.ServerError()
	}
	if version, _ := session.Get(context.SessionKeySessionVersion).(int64); version != user.SessionVersion {
		session.Flush()
		return ctx.Error(http.StatusUnauthorized, "Unauthorized")
	}

	ctx.User = user
	ctx.Map(ctx)
	return nil
}

func clearPendingSignIn(session session.Session) {
	session.Delete(context.SessionKeyPendingUserID)
	session.Delete(context.SessionKeyPendingExpiresAt)
	session.Delete(context.SessionKeyPendingAttempts)
}

func (h *AuthHandler) completeSignIn(ctx context.Context, session session.Session, user *db.User) error {
	clearPendingSignIn(session)
	session.Set(context.SessionKeyUserID, user.ID)
	session.Set(context.SessionKeySessionVersion, user.SessionVersion)

	if err := db.Users.TouchSignIn(ctx.Request().Context(), user.ID); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Warn("Failed to record sign-in time")
	}
	return ctx.Success(dto.SignInResp{Token: session.ID()})
}

// SignIn sign in with a username and password.
// @Summary Sign in with a username and password
// @Tags Auth
// @Description When the user has two-factor authentication enabled, twoFactorRequired is true and the returned token must be sent with signInTwoFactor within 5 minutes.
// @Accept json
// @Produce json
// @Param data body form.SignIn true "Request body"
// @Success 200 {object} dto.Response[dto.SignInResp]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 403 {object} dto.ErrorResponse "Invalid username or password"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID signIn
// @Router /api/auth/sign-in [post]
func (h *AuthHandler) SignIn(ctx context.Context, session session.Session, f form.SignIn) error {
	user, err := db.Users.GetByUsername(ctx.Request().Context(), f.Username)
	if err != nil && !errors.Is(err, db.ErrUserNotFound) {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get user")
		return ctx.ServerError()
	}

	var passwordHash string
	if user != nil {
		passwordHash = user.PasswordHash
	}
	if !auth.CheckPassword(passwordHash, f.Password) {
		return ctx.Error(http.StatusForbidden, "Invalid username or password")
	}

	// Sign out any user of the current session first.
	session.Delete(context.SessionKeyUserID)
	session.Delete(context.SessionKeySessionVersion)

	if user.TOTPEnabled {
		session.Set(context.SessionKeyPendingUserID, user.ID)
		session.Set(context.SessionKeyPendingExpiresAt, time.Now().Add(twoFactorTimeout))
		session.Set(context.SessionKeyPendingAttempts, 0)
		return ctx.Success(dto.SignInResp{Token: session.ID(), TwoFactorRequired: true})
	}
	return h.completeSignIn(ctx, session, user)
}

// SignInTwoFactor complete sign-in with a verification code.
// @Summary Complete sign-in with a verification code
// @Tags Auth
// @Description Accepts a TOTP code or an unused recovery code. After 5 wrong codes or 5 minutes the pending sign-in is discarded and the client must sign in again.
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param data body form.SignInTwoFactor true "Request body"
// @Success 200 {object} dto.Response[dto.SignInResp]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 403 {object} dto.ErrorResponse "Invalid verification code"
// @Failure 410 {object} dto.ErrorResponse "No pending sign-in, or it expired"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID signInTwoFactor
// @Router /api/auth/sign-in/2fa [post]
func (h *AuthHandler) SignInTwoFactor(ctx context.Context, session session.Session, f form.SignInTwoFactor) error {
	userID, _ := session.Get(context.SessionKeyPendingUserID).(uint)
	expiresAt, _ := session.Get(context.SessionKeyPendingExpiresAt).(time.Time)
	if userID == 0 || time.Now().After(expiresAt) {
		clearPendingSignIn(session)
		return ctx.Error(http.StatusGone, "Verification expired. Please sign in again.")
	}

	user, err := db.Users.GetByID(ctx.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			clearPendingSignIn(session)
			return ctx.Error(http.StatusGone, "Verification expired. Please sign in again.")
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to get user")
		return ctx.ServerError()
	}
	if !user.TOTPEnabled {
		// Two-factor was turned off while pending, so finish signing in.
		return h.completeSignIn(ctx, session, user)
	}

	valid := auth.ValidateTOTP(f.Code, user.TOTPSecret)
	if !valid {
		valid, err = db.Users.UseRecoveryCode(ctx.Request().Context(), user.ID, auth.HashRecoveryCode(f.Code))
		if err != nil {
			logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to use recovery code")
			return ctx.ServerError()
		}
	}
	if !valid {
		attempts, _ := session.Get(context.SessionKeyPendingAttempts).(int)
		attempts++
		if attempts >= maxTwoFactorAttempts {
			clearPendingSignIn(session)
			return ctx.Error(http.StatusGone, "Too many incorrect codes. Please sign in again.")
		}
		session.Set(context.SessionKeyPendingAttempts, attempts)
		return ctx.Error(http.StatusForbidden, "Invalid verification code")
	}
	return h.completeSignIn(ctx, session, user)
}

// Profile get the signed-in user.
// @Summary Get the signed-in user
// @Tags Auth
// @Produce json
// @Security SessionAuth
// @Success 200 {object} dto.Response[dto.User]
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @ID profile
// @Router /api/auth/profile [get]
func (h *AuthHandler) Profile(ctx context.Context) error {
	return ctx.Success(dto.ToUser(ctx.User))
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
	return ctx.Success("Signed out")
}
