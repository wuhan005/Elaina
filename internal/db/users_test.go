// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"errors"
	"testing"
)

func TestUsers(t *testing.T) {
	newTestDatabase(t)
	ctx := context.Background()

	created, err := Users.EnsureAdmin(ctx, "hash-1")
	if err != nil || !created {
		t.Fatalf("expected the initial admin to be created: %v %v", created, err)
	}
	if created, err := Users.EnsureAdmin(ctx, "hash-2"); err != nil || created {
		t.Fatalf("initial admin must only be created once: %v %v", created, err)
	}
	admin, err := Users.GetByUsername(ctx, BootstrapUsername)
	if err != nil || admin.PasswordHash != "hash-1" {
		t.Fatalf("unexpected admin: %+v %v", admin, err)
	}

	if _, err := Users.Create(ctx, CreateUserOptions{Username: BootstrapUsername, PasswordHash: "x"}); !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}

	version, err := Users.UpdatePassword(ctx, admin.ID, "hash-3")
	if err != nil || version != admin.SessionVersion+1 {
		t.Fatalf("expected session version %d, got %d (%v)", admin.SessionVersion+1, version, err)
	}

	if err := Users.SetTOTPSecret(ctx, admin.ID, "SECRET"); err != nil {
		t.Fatal(err)
	}
	if used, err := Users.UseRecoveryCode(ctx, admin.ID, "code-a"); err != nil || used {
		t.Fatalf("recovery codes must not work before TOTP is enabled: %v %v", used, err)
	}
	if err := Users.EnableTOTP(ctx, admin.ID, []string{"code-a", "code-b"}); err != nil {
		t.Fatal(err)
	}
	if used, err := Users.UseRecoveryCode(ctx, admin.ID, "code-a"); err != nil || !used {
		t.Fatalf("expected recovery code to be accepted: %v %v", used, err)
	}
	if used, err := Users.UseRecoveryCode(ctx, admin.ID, "code-a"); err != nil || used {
		t.Fatalf("recovery code must be single use: %v %v", used, err)
	}
	admin, err = Users.GetByID(ctx, admin.ID)
	if err != nil || !admin.TOTPEnabled || len(admin.RecoveryCodes) != 1 || admin.RecoveryCodes[0] != "code-b" {
		t.Fatalf("unexpected TOTP state: %+v %v", admin, err)
	}

	before := admin.SessionVersion
	if err := Users.DisableTOTP(ctx, admin.ID, true); err != nil {
		t.Fatal(err)
	}
	admin, err = Users.GetByID(ctx, admin.ID)
	if err != nil || admin.TOTPEnabled || admin.TOTPSecret != "" || len(admin.RecoveryCodes) != 0 || admin.SessionVersion != before+1 {
		t.Fatalf("unexpected state after reset: %+v %v", admin, err)
	}

	if err := Users.TouchSignIn(ctx, admin.ID); err != nil {
		t.Fatal(err)
	}
	if err := Users.Delete(ctx, admin.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := Users.GetByID(ctx, admin.ID); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
