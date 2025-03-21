package handler

import (
	"crypto/subtle"
	"database/sql"
	"net/http"
	"time"

	"github.com/nthnn/ura/logger"
	"github.com/nthnn/ura/util"
)

func authenticate(db *sql.DB, r *http.Request) (*User, string) {
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		return nil, errInvalidLoginCredentials
	}

	if !util.ValidateSessionToken(sessionToken) {
		return nil, errInvalidLoginCredentials
	}

	securityCode := r.Header.Get("X-Security-Code")
	if securityCode == "" {
		return nil, errInvalidLoginCredentials
	}

	if !util.ValidateSecurityCode(securityCode) {
		return nil, errInvalidLoginCredentials
	}

	var userID int64
	var expiresAtStr string

	err := db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE token = ?",
		sessionToken,
	).Scan(&userID, &expiresAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errInvalidLoginCredentials
		}

		logger.Error("Error querying session: %s", err.Error())
		return nil, errInvalidLoginCredentials
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		logger.Error("Error parsing session expiry: %s", err.Error())
		return nil, errInvalidLoginCredentials
	}

	if time.Now().After(expiresAt) {
		return nil, errInvalidLoginCredentials
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
		if err == sql.ErrNoRows {
			return nil, errInvalidLoginCredentials
		}

		logger.Error("Error querying user: %s", err.Error())
		return nil, errInternalErrorOccurred
	}

	user.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		logger.Error("Error parsing user createdAt: %s", err.Error())
		return nil, errInternalErrorOccurred
	}

	if subtle.ConstantTimeCompare([]byte(user.SecurityCode), []byte(securityCode)) != 1 {
		return nil, errInvalidLoginCredentials
	}

	return &user, ""
}
