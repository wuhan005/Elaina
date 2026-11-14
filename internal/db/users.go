// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/dbutil"
)

var _ UserStore = (*users)(nil)

var Users UserStore

type UserStore interface {
	// List returns all the users ordered by ID.
	List(ctx context.Context) ([]*User, error)
	// GetByID returns a user with the given id.
	GetByID(ctx context.Context, id uint) (*User, error)
	// GetByUsername returns a user with the given username.
	GetByUsername(ctx context.Context, username string) (*User, error)
	// Create creates a new user with the given options.
	Create(ctx context.Context, options CreateUserOptions) (*User, error)
	// Delete deletes a user with the given id.
	Delete(ctx context.Context, id uint) error
	// UpdatePassword replaces the password hash and invalidates existing sessions.
	// It returns the new session version.
	UpdatePassword(ctx context.Context, id uint, passwordHash string) (int64, error)
	// SetTOTPSecret stores a pending TOTP secret that is not enabled yet.
	SetTOTPSecret(ctx context.Context, id uint, secret string) error
	// EnableTOTP enables the pending TOTP secret and replaces the recovery codes.
	EnableTOTP(ctx context.Context, id uint, recoveryCodeHashes []string) error
	// DisableTOTP clears the TOTP secret and recovery codes. When invalidateSessions
	// is true, existing sessions of the user are signed out.
	DisableTOTP(ctx context.Context, id uint, invalidateSessions bool) error
	// SetRecoveryCodes replaces the recovery code hashes.
	SetRecoveryCodes(ctx context.Context, id uint, recoveryCodeHashes []string) error
	// UseRecoveryCode consumes a recovery code atomically. It reports whether the
	// code was valid.
	UseRecoveryCode(ctx context.Context, id uint, recoveryCodeHash string) (bool, error)
	// TouchSignIn records a successful sign-in.
	TouchSignIn(ctx context.Context, id uint) error
	// EnsureAdmin creates the "admin" user with the given password hash when the
	// table is empty.
	EnsureAdmin(ctx context.Context, passwordHash string) (bool, error)
}

func NewUserStore(db *gorm.DB) UserStore {
	return &users{db}
}

type users struct {
	*gorm.DB
}

// User is hard-deleted so that usernames stay unique in the table.
type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	// Incrementing SessionVersion invalidates every session issued before.
	SessionVersion int64 `gorm:"not null;default:0"`

	TOTPSecret    string
	TOTPEnabled   bool           `gorm:"not null;default:false"`
	RecoveryCodes pq.StringArray `gorm:"type:text[]"`

	LastSignInAt *time.Time
}

const BootstrapUsername = "admin"

var (
	ErrUserNotFound = errors.New("user does not exist")
	ErrUserExists   = errors.New("username already exists")
)

func (db *users) List(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := db.WithContext(ctx).Order("id ASC").Find(&users).Error; err != nil {
		return nil, errors.Wrap(err, "find")
	}
	return users, nil
}

func (db *users) get(ctx context.Context, query string, args ...any) (*User, error) {
	var user User
	if err := db.WithContext(ctx).Where(query, args...).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrap(err, "first")
	}
	return &user, nil
}

func (db *users) GetByID(ctx context.Context, id uint) (*User, error) {
	return db.get(ctx, "id = ?", id)
}

func (db *users) GetByUsername(ctx context.Context, username string) (*User, error) {
	return db.get(ctx, "username = ?", username)
}

type CreateUserOptions struct {
	Username     string
	PasswordHash string
}

func (db *users) Create(ctx context.Context, options CreateUserOptions) (*User, error) {
	user := &User{
		Username:     options.Username,
		PasswordHash: options.PasswordHash,
	}
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUserExists
		}
		return nil, errors.Wrap(err, "create")
	}
	return user, nil
}

func (db *users) Delete(ctx context.Context, id uint) error {
	result := db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "delete")
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (db *users) update(ctx context.Context, id uint, values map[string]any) error {
	result := db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return errors.Wrap(result.Error, "update")
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (db *users) UpdatePassword(ctx context.Context, id uint, passwordHash string) (int64, error) {
	if err := db.update(ctx, id, map[string]any{
		"password_hash":   passwordHash,
		"session_version": gorm.Expr("session_version + 1"),
	}); err != nil {
		return 0, err
	}
	user, err := db.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.SessionVersion, nil
}

func (db *users) SetTOTPSecret(ctx context.Context, id uint, secret string) error {
	return db.update(ctx, id, map[string]any{
		"totp_secret":    secret,
		"totp_enabled":   false,
		"recovery_codes": pq.StringArray{},
	})
}

func (db *users) EnableTOTP(ctx context.Context, id uint, recoveryCodeHashes []string) error {
	return db.update(ctx, id, map[string]any{
		"totp_enabled":   true,
		"recovery_codes": pq.StringArray(recoveryCodeHashes),
	})
}

func (db *users) DisableTOTP(ctx context.Context, id uint, invalidateSessions bool) error {
	values := map[string]any{
		"totp_secret":    "",
		"totp_enabled":   false,
		"recovery_codes": pq.StringArray{},
	}
	if invalidateSessions {
		values["session_version"] = gorm.Expr("session_version + 1")
	}
	return db.update(ctx, id, values)
}

func (db *users) SetRecoveryCodes(ctx context.Context, id uint, recoveryCodeHashes []string) error {
	return db.update(ctx, id, map[string]any{
		"recovery_codes": pq.StringArray(recoveryCodeHashes),
	})
}

func (db *users) UseRecoveryCode(ctx context.Context, id uint, recoveryCodeHash string) (bool, error) {
	// The conditional update ensures a recovery code is used only once, even under concurrent requests.
	result := db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND totp_enabled AND ? = ANY(recovery_codes)", id, recoveryCodeHash).
		Update("recovery_codes", gorm.Expr("array_remove(recovery_codes, ?)", recoveryCodeHash))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "update")
	}
	return result.RowsAffected == 1, nil
}

func (db *users) TouchSignIn(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&User{}).Where("id = ?", id).
		UpdateColumn("last_sign_in_at", dbutil.Now()).Error
}

func (db *users) EnsureAdmin(ctx context.Context, passwordHash string) (bool, error) {
	var count int64
	if err := db.WithContext(ctx).Model(&User{}).Count(&count).Error; err != nil {
		return false, errors.Wrap(err, "count")
	}
	if count > 0 {
		return false, nil
	}
	if _, err := db.Create(ctx, CreateUserOptions{
		Username:     BootstrapUsername,
		PasswordHash: passwordHash,
	}); err != nil {
		if errors.Is(err, ErrUserExists) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
