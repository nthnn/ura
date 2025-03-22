package util

import (
	"net/mail"
	"regexp"
)

var sha512Regex *regexp.Regexp = regexp.MustCompile("^[a-fA-F0-9]{128}$")
var usernameRegex *regexp.Regexp = regexp.MustCompile(`^[\p{L}\p{N}_.]+$`)

func IsValidSHA512(hash string) bool {
	return len(hash) == 128 && sha512Regex.MatchString(hash)
}

func ValidateUsername(username string) bool {
	return len(username) > 6 && usernameRegex.MatchString(username)
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidateNumbers(s string) bool {
	if s == "" {
		return false
	}

	dotCount := 0
	for _, ch := range s {
		if ch == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

func ValidateSessionToken(s string) bool {
	if len(s) != 64 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

func ValidateSecurityCode(s string) bool {
	if len(s) != 32 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

func ValidateTransactionID(s string) bool {
	if len(s) != 64 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}
