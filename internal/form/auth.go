// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

type SignIn struct {
	// Password must contain 8-72 UTF-8 bytes.
	Password string `json:"password" valid:"required"`
} // @name SignInForm
