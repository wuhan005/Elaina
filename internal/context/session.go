// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"strings"
)

func ReadIDFunc(r *http.Request) string {
	_, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(token)
}

func WriteIDFunc(w http.ResponseWriter, r *http.Request, sid string, created bool) {
	// Do nothing.
}
