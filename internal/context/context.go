// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

// Context represents context of a request.
type Context struct {
	flamego.Context
	IsAuthenticated bool
}

func (c *Context) Success(data ...interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.ResponseWriter().WriteHeader(http.StatusOK)

	var d interface{}
	if len(data) == 1 {
		d = data[0]
	}

	err := json.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"data": d,
		},
	)
	if err != nil {
		log.Error("Failed to encode", "error", err)
	}
	return nil
}

func (c *Context) ServerError() error {
	return c.Error(http.StatusInternalServerError, "internal server error")
}

func (c *Context) Error(statusCode int, message string, v ...interface{}) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.ResponseWriter().WriteHeader(statusCode)

	err := json.NewEncoder(c.ResponseWriter()).Encode(
		map[string]interface{}{
			"msg": fmt.Sprintf(message, v...),
		},
	)
	if err != nil {
		log.Error("Failed to encode", "error", err)
	}
	return nil
}

func (c *Context) Status(statusCode int) {
	c.ResponseWriter().WriteHeader(statusCode)
}

const (
	SessionIDIsAuthenticated = "_isAuthenticated"
)

// Contexter initializes a classic context for a request.
func Contexter(gormDB *gorm.DB) flamego.Handler {
	return func(ctx flamego.Context, session session.Session) {
		c := Context{
			Context:         ctx,
			IsAuthenticated: false,
		}

		isAuthenticatedInf := session.Get(SessionIDIsAuthenticated)
		if isAuthenticated, ok := isAuthenticatedInf.(bool); ok && isAuthenticated {
			c.IsAuthenticated = true
		}

		c.MapTo(gormDB, (*dbutil.Transactor)(nil))
		c.Map(c)
	}
}
