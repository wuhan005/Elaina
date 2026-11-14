// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/flamego/flamego"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/wuhan005/Elaina/internal/auth"
	"github.com/wuhan005/Elaina/internal/db"
)

type testClient struct {
	t      *testing.T
	router *flamego.Flame
}

func newTestClient(t *testing.T) *testClient {
	t.Helper()
	router, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return &testClient{t: t, router: router}
}

// do sends a request and returns the status code and the decoded body.
func (c *testClient) do(method, path, token string, body any) (int, map[string]any) {
	c.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Token "+token)
	}
	response := httptest.NewRecorder()
	c.router.ServeHTTP(response, request)

	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		c.t.Fatalf("invalid JSON response %q: %v", response.Body.String(), err)
	}
	if response.Code != http.StatusOK {
		if msg, _ := result["msg"].(string); len(result) != 1 || msg == "" {
			c.t.Fatalf("unexpected error envelope: %s", response.Body.String())
		}
	}
	return response.Code, result
}

func (c *testClient) signIn(username, password string) (int, string, bool) {
	c.t.Helper()
	status, result := c.do(http.MethodPost, "/api/auth/sign-in", "", map[string]string{"username": username, "password": password})
	if status != http.StatusOK {
		return status, "", false
	}
	data := result["data"].(map[string]any)
	return status, data["token"].(string), data["twoFactorRequired"].(bool)
}

func createTestUser(t *testing.T, store *fakeUsers, username, password string) *db.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.Create(t.Context(), db.CreateUserOptions{Username: username, PasswordHash: string(hash)})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestSignIn(t *testing.T) {
	store := useFakeUsers(t)
	createTestUser(t, store, "admin", "admin-password")
	client := newTestClient(t)

	for _, test := range []struct {
		name, username, password string
		status                   int
	}{
		{"correct password", "admin", "admin-password", http.StatusOK},
		{"wrong password", "admin", "wrong-password", http.StatusForbidden},
		{"unknown user", "nobody", "admin-password", http.StatusForbidden},
		{"username is case sensitive", "Admin", "admin-password", http.StatusForbidden},
		{"empty password", "admin", "", http.StatusBadRequest},
		{"empty username", "", "admin-password", http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			status, token, twoFactor := client.signIn(test.username, test.password)
			if status != test.status {
				t.Fatalf("expected status %d, got %d", test.status, status)
			}
			if twoFactor {
				t.Fatal("two-factor should not be required")
			}

			profileStatus, profile := client.do(http.MethodGet, "/api/auth/profile", token, nil)
			if test.status != http.StatusOK {
				if profileStatus != http.StatusUnauthorized {
					t.Fatalf("expected profile status 401, got %d", profileStatus)
				}
				return
			}
			if profileStatus != http.StatusOK {
				t.Fatalf("expected profile status 200, got %d", profileStatus)
			}
			data := profile["data"].(map[string]any)
			if data["username"] != "admin" || data["twoFactorEnabled"] != false || data["lastSignInAt"] == nil {
				t.Fatalf("unexpected profile: %v", data)
			}
		})
	}
}

func TestLocalizedMessages(t *testing.T) {
	store := useFakeUsers(t)
	createTestUser(t, store, "admin", "admin-password")
	client := newTestClient(t)

	request := func(query, language string) string {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in"+query, strings.NewReader(`{"username":"admin","password":"wrong-password"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept-Language", language)
		w := httptest.NewRecorder()
		client.router.ServeHTTP(w, r)
		var body struct{ Msg string }
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Msg
	}
	for _, test := range []struct{ query, language, want string }{
		{"", "zh-CN,zh;q=0.9", "用户名或密码错误"},
		{"", "en-US", "Invalid username or password"},
		{"?lang=zh-CN", "en-US", "用户名或密码错误"},
		{"?lang=en-US", "zh-CN", "Invalid username or password"},
	} {
		if got := request(test.query, test.language); got != test.want {
			t.Fatalf("query %q, Accept-Language %q: got %q, want %q", test.query, test.language, got, test.want)
		}
	}
}

func TestSignInTwoFactor(t *testing.T) {
	store := useFakeUsers(t)
	createTestUser(t, store, "admin", "admin-password")
	client := newTestClient(t)

	_, token, _ := client.signIn("admin", "admin-password")
	status, setup := client.do(http.MethodPost, "/api/account/2fa/setup", token, nil)
	if status != http.StatusOK {
		t.Fatalf("setup: expected 200, got %d", status)
	}
	secret := setup["data"].(map[string]any)["secret"].(string)

	if status, _ := client.do(http.MethodPost, "/api/account/2fa/enable", token, map[string]string{"code": "000000"}); status != http.StatusForbidden {
		t.Fatalf("enable with wrong code: expected 403, got %d", status)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	status, enabled := client.do(http.MethodPost, "/api/account/2fa/enable", token, map[string]string{"code": code})
	if status != http.StatusOK {
		t.Fatalf("enable: expected 200, got %d", status)
	}
	recoveryCodes := enabled["data"].(map[string]any)["codes"].([]any)
	if len(recoveryCodes) != auth.RecoveryCodeCount {
		t.Fatalf("expected %d recovery codes, got %d", auth.RecoveryCodeCount, len(recoveryCodes))
	}

	// With two-factor enabled, a password sign-in only returns a pending token.
	_, pending, twoFactor := client.signIn("admin", "admin-password")
	if !twoFactor {
		t.Fatal("expected two-factor to be required")
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", pending, nil); status != http.StatusUnauthorized {
		t.Fatalf("pending token must not authenticate, got %d", status)
	}
	if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", pending, map[string]string{"code": "123456"}); status != http.StatusForbidden {
		t.Fatalf("wrong code: expected 403, got %d", status)
	}
	if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", pending, map[string]string{"code": code}); status != http.StatusOK {
		t.Fatalf("correct code: expected 200, got %d", status)
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", pending, nil); status != http.StatusOK {
		t.Fatalf("verified token should authenticate, got %d", status)
	}

	t.Run("recovery code is single use", func(t *testing.T) {
		recovery := strings.ToUpper(recoveryCodes[0].(string))
		_, first, _ := client.signIn("admin", "admin-password")
		if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", first, map[string]string{"code": recovery}); status != http.StatusOK {
			t.Fatalf("recovery code: expected 200, got %d", status)
		}
		_, second, _ := client.signIn("admin", "admin-password")
		if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", second, map[string]string{"code": recovery}); status != http.StatusForbidden {
			t.Fatalf("reused recovery code: expected 403, got %d", status)
		}
		_, profile := client.do(http.MethodGet, "/api/auth/profile", first, nil)
		if remaining := profile["data"].(map[string]any)["recoveryCodesRemaining"]; remaining != float64(auth.RecoveryCodeCount-1) {
			t.Fatalf("expected %d recovery codes left, got %v", auth.RecoveryCodeCount-1, remaining)
		}
	})

	t.Run("too many attempts", func(t *testing.T) {
		_, pending, _ := client.signIn("admin", "admin-password")
		for i := 1; i < maxTwoFactorAttempts; i++ {
			if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", pending, map[string]string{"code": "000000"}); status != http.StatusForbidden {
				t.Fatalf("attempt %d: expected 403, got %d", i, status)
			}
		}
		if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", pending, map[string]string{"code": "000000"}); status != http.StatusGone {
			t.Fatalf("last attempt: expected 410, got %d", status)
		}
		if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", pending, map[string]string{"code": code}); status != http.StatusGone {
			t.Fatalf("discarded sign-in: expected 410, got %d", status)
		}
	})

	t.Run("no pending sign-in", func(t *testing.T) {
		if status, _ := client.do(http.MethodPost, "/api/auth/sign-in/2fa", "", map[string]string{"code": code}); status != http.StatusGone {
			t.Fatalf("expected 410, got %d", status)
		}
	})

	t.Run("disable requires password", func(t *testing.T) {
		if status, _ := client.do(http.MethodPost, "/api/account/2fa/disable", token, map[string]string{"password": "wrong-password"}); status != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", status)
		}
		if status, _ := client.do(http.MethodPost, "/api/account/2fa/disable", token, map[string]string{"password": "admin-password"}); status != http.StatusOK {
			t.Fatalf("expected 200, got %d", status)
		}
		if _, _, twoFactor := client.signIn("admin", "admin-password"); twoFactor {
			t.Fatal("two-factor should be disabled")
		}
	})
}

func TestSessionInvalidation(t *testing.T) {
	store := useFakeUsers(t)
	createTestUser(t, store, "admin", "admin-password")
	member := createTestUser(t, store, "member", "member-password")
	client := newTestClient(t)

	_, admin, _ := client.signIn("admin", "admin-password")
	_, otherAdminSession, _ := client.signIn("admin", "admin-password")
	_, memberSession, _ := client.signIn("member", "member-password")

	// Changing your own password keeps the current session and signs out the others.
	if status, _ := client.do(http.MethodPut, "/api/account/password", admin, map[string]string{"currentPassword": "wrong-password", "newPassword": "new-password"}); status != http.StatusForbidden {
		t.Fatalf("wrong current password: expected 403, got %d", status)
	}
	if status, _ := client.do(http.MethodPut, "/api/account/password", admin, map[string]string{"currentPassword": "admin-password", "newPassword": "short"}); status != http.StatusBadRequest {
		t.Fatalf("short new password: expected 400, got %d", status)
	}
	if status, _ := client.do(http.MethodPut, "/api/account/password", admin, map[string]string{"currentPassword": "admin-password", "newPassword": "new-password"}); status != http.StatusOK {
		t.Fatalf("change password: expected 200, got %d", status)
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", admin, nil); status != http.StatusOK {
		t.Fatalf("current session should stay valid, got %d", status)
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", otherAdminSession, nil); status != http.StatusUnauthorized {
		t.Fatalf("other session should be signed out, got %d", status)
	}

	// Resetting another user's password signs that user out.
	if status, _ := client.do(http.MethodPut, "/api/user/"+itoa(member.ID)+"/password", admin, map[string]string{"password": "reset-password"}); status != http.StatusOK {
		t.Fatalf("reset password: expected 200, got %d", status)
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", memberSession, nil); status != http.StatusUnauthorized {
		t.Fatalf("member session should be signed out, got %d", status)
	}
	if status, _, _ := client.signIn("member", "reset-password"); status != http.StatusOK {
		t.Fatalf("member should sign in with the new password, got %d", status)
	}
}

func TestUserManagement(t *testing.T) {
	store := useFakeUsers(t)
	self := createTestUser(t, store, "admin", "admin-password")
	client := newTestClient(t)
	_, token, _ := client.signIn("admin", "admin-password")

	for _, test := range []struct {
		name               string
		username, password string
		status             int
	}{
		{"valid user", "member", "member-password", http.StatusOK},
		{"duplicate username", "member", "member-password", http.StatusConflict},
		{"invalid username", "-bad", "member-password", http.StatusBadRequest},
		{"short password", "member2", "short", http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			status, _ := client.do(http.MethodPost, "/api/users", token, map[string]string{"username": test.username, "password": test.password})
			if status != test.status {
				t.Fatalf("expected status %d, got %d", test.status, status)
			}
		})
	}

	status, list := client.do(http.MethodGet, "/api/users", token, nil)
	if status != http.StatusOK || len(list["data"].([]any)) != 2 {
		t.Fatalf("unexpected user list: %d %v", status, list)
	}
	if status, _ := client.do(http.MethodDelete, "/api/user/"+itoa(self.ID), token, nil); status != http.StatusBadRequest {
		t.Fatalf("deleting yourself: expected 400, got %d", status)
	}
	member, _ := store.GetByUsername(t.Context(), "member")
	if status, _ := client.do(http.MethodDelete, "/api/user/"+itoa(member.ID), token, nil); status != http.StatusOK {
		t.Fatalf("delete member: expected 200, got %d", status)
	}
	if status, _ := client.do(http.MethodDelete, "/api/user/"+itoa(member.ID), token, nil); status != http.StatusNotFound {
		t.Fatalf("delete missing member: expected 404, got %d", status)
	}

	// Deleting the signed-in user invalidates the session immediately.
	if err := store.Delete(t.Context(), self.ID); err != nil {
		t.Fatal(err)
	}
	if status, _ := client.do(http.MethodGet, "/api/auth/profile", token, nil); status != http.StatusUnauthorized {
		t.Fatalf("deleted user: expected 401, got %d", status)
	}
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
