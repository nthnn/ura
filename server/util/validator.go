package util

import (
	"net/mail"
	"regexp"
)

func IsValidSHA512(hash string) bool {
	if len(hash) != 128 {
		return false
	}

	match, _ := regexp.MatchString("^[a-fA-F0-9]{128}$", hash)
	return match
}

func ValidateUsername(username string) bool {
	re := regexp.MustCompile(`^[\p{L}\p{N}_.]+$`)
	return re.MatchString(username)
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidateNumbers(s string) bool {
	for _, ch := range s {
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
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
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
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
