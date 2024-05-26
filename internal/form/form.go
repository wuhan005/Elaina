// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/flamego/flamego"
	"github.com/sirupsen/logrus"
	"github.com/wuhan005/govalid"
	"golang.org/x/text/language"
)

type ErrorCategory string

const (
	ErrorCategoryDeserialization ErrorCategory = "deserialization"
	ErrorCategoryValidation      ErrorCategory = "validation"
)

type Error struct {
	Category ErrorCategory
	Error    error
}

func Bind(model interface{}) flamego.Handler {
	// Ensure not pointer.
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		panic("form: pointer can not be accepted as binding model")
	}

	return flamego.ContextInvoker(func(c flamego.Context) {
		obj := reflect.New(reflect.TypeOf(model))
		r := c.Request().Request
		if r.Body != nil {
			defer func() { _ = r.Body.Close() }()

			if err := json.NewDecoder(r.Body).Decode(obj.Interface()); err != nil {
				c.Map(Error{Category: ErrorCategoryDeserialization, Error: err})
				if _, err := c.Invoke(errorHandler); err != nil {
					panic("form: " + err.Error())
				}
				return
			}
		}

		acceptLanguage := r.Header.Get("Accept-Language")
		languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		languageTag := language.Chinese
		if len(languageTags) > 0 {
			languageTag = languageTags[0]
		}

		errors, ok := govalid.Check(obj.Interface(), languageTag)
		if !ok {
			c.Map(Error{Category: ErrorCategoryValidation, Error: errors[0]})
			if _, err := c.Invoke(errorHandler); err != nil {
				panic("form: " + err.Error())
			}
			return
		}

		// Validation passed.
		c.Map(obj.Elem().Interface())
	})
}

func errorHandler(c flamego.Context, error Error) {
	c.ResponseWriter().WriteHeader(http.StatusBadRequest)
	c.ResponseWriter().Header().Set("Content-Type", "application/json; charset=utf-8")

	var msg string
	if error.Category == ErrorCategoryDeserialization {
		msg = "invalid request body"
	} else {
		msg = error.Error.Error()
	}

	body := map[string]interface{}{
		"msg": msg,
	}
	err := json.NewEncoder(c.ResponseWriter()).Encode(body)
	if err != nil {
		logrus.WithError(err).Error("Failed to encode response")
	}
}
