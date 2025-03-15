package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/nthnn/ura/util"
)

func authenticate(db *sql.DB, r *http.Request) (*User, error) {
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		return nil, errors.New("missing session token")
	}

	if !util.ValidateSessionToken(sessionToken) {
		return nil, errors.New("invalid session token")
	}

	securityCode := r.Header.Get("X-Security-Code")
	if securityCode == "" {
		return nil, errors.New("missing security code")
	}

	if !util.ValidateSecurityCode(securityCode) {
		return nil, errors.New("invalid security code")
	}

	var userID int64
	var expiresAtStr string

	err := db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE token = ?",
		sessionToken,
	).Scan(&userID, &expiresAtStr)
	if err != nil {
		return nil, errors.New("invalid session token")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, errors.New("session parse error")
	}

	if time.Now().After(expiresAt) {
		return nil, errors.New("session expired")
	}

	var user User
	var createdAtStr string

	err = db.QueryRow(
		"SELECT id, username, email, identifier, security_code, balance_ura, created_at FROM users WHERE id = ?",
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Identifier,
		&user.SecurityCode,
		&user.BalanceUra,
		&createdAtStr,
	)

	if err != nil {
		return nil, errors.New("user not found")
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	if user.SecurityCode != securityCode {
		return nil, errors.New("invalid security code")
	}

	return &user, nil
}
