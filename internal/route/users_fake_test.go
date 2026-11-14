// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/wuhan005/Elaina/internal/db"
)

// fakeUsers is an in-memory UserStore for route tests.
type fakeUsers struct {
	mu     sync.Mutex
	nextID uint
	users  map[uint]*db.User
}

var _ db.UserStore = (*fakeUsers)(nil)

func useFakeUsers(t *testing.T) *fakeUsers {
	t.Helper()
	store := &fakeUsers{nextID: 1, users: map[uint]*db.User{}}
	previous := db.Users
	db.Users = store
	t.Cleanup(func() { db.Users = previous })
	return store
}

func (s *fakeUsers) copyOf(user *db.User) *db.User {
	clone := *user
	clone.RecoveryCodes = slices.Clone(user.RecoveryCodes)
	return &clone
}

func (s *fakeUsers) List(context.Context) ([]*db.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	users := make([]*db.User, 0, len(s.users))
	for id := uint(1); id < s.nextID; id++ {
		if user, ok := s.users[id]; ok {
			users = append(users, s.copyOf(user))
		}
	}
	return users, nil
}

func (s *fakeUsers) GetByID(_ context.Context, id uint) (*db.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return nil, db.ErrUserNotFound
	}
	return s.copyOf(user), nil
}

func (s *fakeUsers) GetByUsername(_ context.Context, username string) (*db.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.Username == username {
			return s.copyOf(user), nil
		}
	}
	return nil, db.ErrUserNotFound
}

func (s *fakeUsers) Create(_ context.Context, options db.CreateUserOptions) (*db.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.Username == options.Username {
			return nil, db.ErrUserExists
		}
	}
	user := &db.User{ID: s.nextID, Username: options.Username, PasswordHash: options.PasswordHash, CreatedAt: time.Now()}
	s.users[user.ID] = user
	s.nextID++
	return s.copyOf(user), nil
}

func (s *fakeUsers) Delete(_ context.Context, id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return db.ErrUserNotFound
	}
	delete(s.users, id)
	return nil
}

func (s *fakeUsers) modify(id uint, fn func(*db.User)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return db.ErrUserNotFound
	}
	fn(user)
	return nil
}

func (s *fakeUsers) UpdatePassword(_ context.Context, id uint, passwordHash string) (int64, error) {
	var version int64
	err := s.modify(id, func(user *db.User) {
		user.PasswordHash = passwordHash
		user.SessionVersion++
		version = user.SessionVersion
	})
	return version, err
}

func (s *fakeUsers) SetTOTPSecret(_ context.Context, id uint, secret string) error {
	return s.modify(id, func(user *db.User) {
		user.TOTPSecret, user.TOTPEnabled, user.RecoveryCodes = secret, false, nil
	})
}

func (s *fakeUsers) EnableTOTP(_ context.Context, id uint, hashes []string) error {
	return s.modify(id, func(user *db.User) {
		user.TOTPEnabled, user.RecoveryCodes = true, slices.Clone(hashes)
	})
}

func (s *fakeUsers) DisableTOTP(_ context.Context, id uint, invalidateSessions bool) error {
	return s.modify(id, func(user *db.User) {
		user.TOTPSecret, user.TOTPEnabled, user.RecoveryCodes = "", false, nil
		if invalidateSessions {
			user.SessionVersion++
		}
	})
}

func (s *fakeUsers) SetRecoveryCodes(_ context.Context, id uint, hashes []string) error {
	return s.modify(id, func(user *db.User) { user.RecoveryCodes = slices.Clone(hashes) })
}

func (s *fakeUsers) UseRecoveryCode(_ context.Context, id uint, hash string) (bool, error) {
	used := false
	err := s.modify(id, func(user *db.User) {
		if index := slices.Index(user.RecoveryCodes, hash); user.TOTPEnabled && index >= 0 {
			user.RecoveryCodes = slices.Delete(user.RecoveryCodes, index, index+1)
			used = true
		}
	})
	return used, err
}

func (s *fakeUsers) TouchSignIn(_ context.Context, id uint) error {
	now := time.Now()
	return s.modify(id, func(user *db.User) { user.LastSignInAt = &now })
}

func (s *fakeUsers) EnsureAdmin(ctx context.Context, passwordHash string) (bool, error) {
	if users, _ := s.List(ctx); len(users) > 0 {
		return false, nil
	}
	_, err := s.Create(ctx, db.CreateUserOptions{Username: db.BootstrapUsername, PasswordHash: passwordHash})
	return err == nil, err
}
