// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"strings"
)

func ReadIDFunc(r *http.Request) string {
	authorizationHeader := r.Header.Get("Authorization")
	groups := strings.SplitN(authorizationHeader, " ", 2)
	if len(groups) != 2 {
		return ""
	}
	return strings.TrimSpace(groups[1])
}

func WriteIDFunc(w http.ResponseWriter, r *http.Request, sid string, created bool) {
	// Do nothing.
}
