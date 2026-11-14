// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/flamego/flamego"
	"github.com/wuhan005/govalid"
	"golang.org/x/text/language"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/i18n"
)

func Bind(model any) flamego.Handler {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Pointer {
		panic("form: pointer can not be accepted as binding model")
	}

	return func(ctx context.Context) {
		obj := reflect.New(typ)
		r := ctx.Request().Request
		if r.Body != nil {
			defer func() { _ = r.Body.Close() }()

			if err := json.NewDecoder(r.Body).Decode(obj.Interface()); err != nil {
				_ = ctx.Error(http.StatusBadRequest, "common::invalid_body")
				return
			}
		}

		lang := language.English
		if i18n.Chinese(ctx.Locale) {
			lang = language.Chinese
		}
		if errs, ok := govalid.Check(obj.Interface(), lang); !ok {
			_ = ctx.ErrorFrom(http.StatusBadRequest, errs[0])
			return
		}
		if validator, ok := obj.Interface().(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				_ = ctx.ErrorFrom(http.StatusBadRequest, err)
				return
			}
		}

		// Validation passed.
		ctx.Map(obj.Elem().Interface())
	}
}
