// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckPassword(t *testing.T) {
	hash := func(password string) string {
		t.Helper()
		value, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(value)
	}
	valid := hash("admin-password")
	long := strings.Repeat("a", MaxPasswordBytes)
	longHash := hash(long)

	for _, test := range []struct {
		name, hash, password string
		want                 bool
	}{
		{"correct password", valid, "admin-password", true},
		{"bcrypt 2b", strings.Replace(valid, "$2a$", "$2b$", 1), "admin-password", true},
		{"bcrypt 2y", strings.Replace(valid, "$2a$", "$2y$", 1), "admin-password", true},
		{"wrong password", valid, "wrong-password", false},
		{"hash is not a password", valid, valid, false},
		{"plaintext hash", "admin-password", "admin-password", false},
		{"missing hash", "", "admin-password", false},
		{"malformed hash", "$2a$04$" + strings.Repeat("!", 53), "admin-password", false},
		{"minimum length", hash("12345678"), "12345678", true},
		{"too short", hash("1234567"), "1234567", false},
		{"maximum length", longHash, long, true},
		{"overlong with matching prefix", longHash, long + "b", false},
		{"multibyte overlong", hash(strings.Repeat("密", 24)), strings.Repeat("密", 25), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := CheckPassword(test.hash, test.password); got != test.want {
				t.Fatalf("CheckPassword() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	for _, name := range []string{"admin", "a1", "john.doe", "ops_team-2", strings.Repeat("a", 32)} {
		if err := ValidateUsername(name); err != nil {
			t.Errorf("%q should be valid: %v", name, err)
		}
	}
	for _, name := range []string{"", "a", "-admin", ".admin", "has space", "中文名", strings.Repeat("a", 33)} {
		if err := ValidateUsername(name); err == nil {
			t.Errorf("%q should be invalid", name)
		}
	}
}

func TestTOTP(t *testing.T) {
	key, err := GenerateTOTP("admin")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key.QRCode, "data:image/png;base64,") || !strings.Contains(key.URL, "issuer=Elaina") {
		t.Fatalf("unexpected key: %+v", key)
	}

	now := time.Unix(1_800_000_000, 0)
	code, err := totp.GenerateCode(key.Secret, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		code string
		at   time.Time
		want bool
	}{
		{"current window", code, now, true},
		{"previous window", code, now.Add(30 * time.Second), true},
		{"spaced code", code[:3] + " " + code[3:], now, true},
		{"expired", code, now.Add(2 * time.Minute), false},
		{"wrong length", code[:5], now, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := validateTOTPAt(test.code, key.Secret, test.at); got != test.want {
				t.Fatalf("validateTOTPAt() = %v, want %v", got, test.want)
			}
		})
	}
	if validateTOTPAt(code, "", now) {
		t.Fatal("empty secret must not validate")
	}
}

func TestRecoveryCodes(t *testing.T) {
	codes, hashes, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != RecoveryCodeCount || len(hashes) != RecoveryCodeCount {
		t.Fatalf("expected %d codes, got %d/%d", RecoveryCodeCount, len(codes), len(hashes))
	}
	seen := map[string]bool{}
	for i, code := range codes {
		if len(code) != 11 || code[5] != '-' {
			t.Fatalf("unexpected code format %q", code)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
		if HashRecoveryCode(code) != hashes[i] {
			t.Fatal("hash mismatch")
		}
		if HashRecoveryCode(" "+strings.ToUpper(strings.ReplaceAll(code, "-", ""))+" ") != hashes[i] {
			t.Fatal("hash should ignore case, spaces and dashes")
		}
	}
}
