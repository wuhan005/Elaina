// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
)

// Context represents context of a request.
type Context struct {
	flamego.Context

	// UserID is the fully signed-in user (including two-factor verification), or 0.
	UserID          uint
	IsAuthenticated bool
	// User is loaded by Authenticator and only set on routes that require sign-in.
	User *db.User
}

// WriteJSON writes v as a JSON response body; encoding errors are only logged.
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logrus.WithContext(r.Context()).WithError(err).Error("Failed to encode")
	}
}

func (c *Context) Success(data ...any) error {
	var d any
	if len(data) == 1 {
		d = data[0]
	}
	WriteJSON(c.ResponseWriter(), c.Request().Request, http.StatusOK, dto.Response[any]{
		Data: d,
	})
	return nil
}

func (c *Context) ServerError() error {
	return c.Error(http.StatusInternalServerError, "internal server error")
}

func (c *Context) Error(statusCode int, message string, v ...any) error {
	WriteJSON(c.ResponseWriter(), c.Request().Request, statusCode, dto.ErrorResponse{
		Msg: fmt.Sprintf(message, v...),
	})
	return nil
}

func (c *Context) Status(statusCode int) {
	c.ResponseWriter().WriteHeader(statusCode)
}

const (
	SessionKeyUserID         = "elaina:user_id"
	SessionKeySessionVersion = "elaina:session_version"

	// Keys used after the password is verified while two-factor verification is pending.
	SessionKeyPendingUserID    = "elaina:pending_user_id"
	SessionKeyPendingExpiresAt = "elaina:pending_expires_at"
	SessionKeyPendingAttempts  = "elaina:pending_attempts"
)

// Contexter initializes a classic context for a request.
func Contexter() flamego.Handler {
	return func(ctx flamego.Context, session session.Session) {
		userID, _ := session.Get(SessionKeyUserID).(uint)
		ctx.Map(Context{
			Context:         ctx,
			UserID:          userID,
			IsAuthenticated: userID != 0,
		})
	}
}
