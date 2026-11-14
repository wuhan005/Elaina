// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package auth provides helpers for passwords, TOTP and recovery codes.
package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"image/png"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordBytes = 8
	// MaxPasswordBytes is the bcrypt limit; longer input would be silently truncated.
	MaxPasswordBytes = 72

	TOTPIssuer         = "Elaina"
	RecoveryCodeCount  = 10
	recoveryCodeLength = 10
)

// timingHash keeps a bcrypt comparison running for unknown users so that response
// times do not reveal which usernames exist.
const timingHash = "$2a$10$3rZTxM59i03M0M8lLZqTvOzNXkYxOvAh7w2S1VLH1AYPFy4YMlg/q"

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,31}$`)

// ValidateUsername requires 2-32 letters, digits, '_', '.' or '-', starting with a
// letter or digit.
func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("Username must be 2-32 characters of letters, digits, '_', '.' or '-', starting with a letter or digit")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < MinPasswordBytes || len(password) > MaxPasswordBytes {
		return errors.Errorf("Password must contain %d-%d UTF-8 bytes", MinPasswordBytes, MaxPasswordBytes)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "generate bcrypt hash")
	}
	return string(hash), nil
}

// CheckPassword reports whether password matches hash. An empty hash is compared with
// a fixed hash to keep the timing the same, and always fails.
func CheckPassword(hash, password string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword([]byte(timingHash), []byte(password))
		return false
	}
	if ValidatePassword(password) != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type TOTPKey struct {
	Secret string
	URL    string
	// QRCode is a PNG data URL usable as <img src>.
	QRCode string
}

func GenerateTOTP(username string) (*TOTPKey, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TOTPIssuer,
		AccountName: username,
	})
	if err != nil {
		return nil, errors.Wrap(err, "generate TOTP key")
	}

	img, err := key.Image(240, 240)
	if err != nil {
		return nil, errors.Wrap(err, "render QR code")
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errors.Wrap(err, "encode QR code")
	}

	return &TOTPKey{
		Secret: key.Secret(),
		URL:    key.URL(),
		QRCode: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// ValidateTOTP allows one 30-second step of clock skew in either direction.
func ValidateTOTP(code, secret string) bool {
	return validateTOTPAt(code, secret, time.Now())
}

func validateTOTPAt(code, secret string, at time.Time) bool {
	code = strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	if secret == "" || len(code) != 6 {
		return false
	}
	ok, err := totp.ValidateCustom(code, secret, at, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return err == nil && ok
}

// recoveryAlphabet leaves out the look-alike characters 0/o and 1/l/i.
const recoveryAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// GenerateRecoveryCodes returns plaintext recovery codes (like abcde-fghjk) and their
// hashes. Only the hashes are stored.
func GenerateRecoveryCodes() (codes, hashes []string, err error) {
	codes = make([]string, 0, RecoveryCodeCount)
	hashes = make([]string, 0, RecoveryCodeCount)
	for range RecoveryCodeCount {
		raw := make([]byte, recoveryCodeLength)
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, errors.Wrap(err, "read random bytes")
		}
		for i, b := range raw {
			raw[i] = recoveryAlphabet[int(b)%len(recoveryAlphabet)]
		}
		code := string(raw[:recoveryCodeLength/2]) + "-" + string(raw[recoveryCodeLength/2:])
		codes = append(codes, code)
		hashes = append(hashes, HashRecoveryCode(code))
	}
	return codes, hashes, nil
}

// HashRecoveryCode hashes the code with SHA-256, ignoring case, whitespace and dashes.
// The codes have enough entropy that a slow hash is not needed.
func HashRecoveryCode(code string) string {
	normalized := strings.Map(func(r rune) rune {
		switch r {
		case '-', ' ', '\t':
			return -1
		}
		return r
	}, strings.ToLower(strings.TrimSpace(code)))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
