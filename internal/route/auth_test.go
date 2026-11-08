// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/form"
)

func TestSignIn(t *testing.T) {
	passwordHash := config.App.PasswordHash
	t.Cleanup(func() { config.App.PasswordHash = passwordHash })
	hash := func(password string) string {
		t.Helper()
		value, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(value)
	}
	validHash := hash("admin-password")
	longPassword := strings.Repeat("a", 72)
	longHash := hash(longPassword)

	router, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, hash, password string
		status               int
	}{
		{"correct password", validHash, "admin-password", http.StatusOK},
		{"bcrypt 2b", strings.Replace(validHash, "$2a$", "$2b$", 1), "admin-password", http.StatusOK},
		{"bcrypt 2y", strings.Replace(validHash, "$2a$", "$2y$", 1), "admin-password", http.StatusOK},
		{"wrong password", validHash, "wrong-password", http.StatusForbidden},
		{"hash is not a password", validHash, validHash, http.StatusForbidden},
		{"plaintext configuration", "admin-password", "admin-password", http.StatusForbidden},
		{"missing hash", "", "admin-password", http.StatusForbidden},
		{"malformed hash", "$2a$04$" + strings.Repeat("!", 53), "admin-password", http.StatusForbidden},
		{"minimum password length", hash("12345678"), "12345678", http.StatusOK},
		{"short password", hash("1234567"), "1234567", http.StatusForbidden},
		{"empty password", hash(""), "", http.StatusBadRequest},
		{"maximum password length", longHash, longPassword, http.StatusOK},
		{"overlong password with matching prefix", longHash, longPassword + "b", http.StatusForbidden},
		{"multibyte password", hash(strings.Repeat("密", 24)), strings.Repeat("密", 25), http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			config.App.PasswordHash = test.hash
			body, err := json.Marshal(form.SignIn{Password: test.password})
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("expected status %d, got %d", test.status, response.Code)
			}
			var result struct{ Data string }
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if (result.Data != "") != (test.status == http.StatusOK) {
				t.Fatal("unexpected session token")
			}

			profile := httptest.NewRecorder()
			request = httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
			request.Header.Set("Authorization", "Bearer "+result.Data)
			router.ServeHTTP(profile, request)
			wantStatus := http.StatusUnauthorized
			if test.status == http.StatusOK {
				wantStatus = http.StatusOK
			}
			if profile.Code != wantStatus {
				t.Fatalf("expected profile status %d, got %d", wantStatus, profile.Code)
			}
		})
	}
}
