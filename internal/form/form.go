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

	"github.com/wuhan005/Elaina/internal/context"
)

func Bind(model interface{}) flamego.Handler {
	// Ensure not pointer.
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		panic("form: pointer can not be accepted as binding model")
	}

	return func(ctx context.Context) {
		obj := reflect.New(reflect.TypeOf(model))
		r := ctx.Request().Request
		if r.Body != nil {
			defer func() { _ = r.Body.Close() }()

			if err := json.NewDecoder(r.Body).Decode(obj.Interface()); err != nil {
				_ = ctx.Error(http.StatusBadRequest, "Failed to parse request body")
				return
			}
		}

		errors, ok := govalid.Check(obj.Interface())
		if !ok {
			_ = ctx.Error(http.StatusBadRequest, errors[0].Error())
			return
		}

		// Validation passed.
		ctx.Map(obj.Elem().Interface())
	}
}
