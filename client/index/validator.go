//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var usernameRegex *regexp.Regexp = regexp.MustCompile(`^[\p{L}\p{N}_.]+$`)

func validateUsername(username string) bool {
	return len(username) > 6 && usernameRegex.MatchString(username)
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func validatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	var repeatedCount int
	var lastRune rune

	for i, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}

		if i > 0 {
			if char == lastRune {
				repeatedCount++
				if repeatedCount >= 2 {
					return errors.New("password should not contain more than 2 identical characters in a row")
				}
			} else {
				repeatedCount = 0
			}
		}

		lastRune = char
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	commonPatterns := []string{"password", "123456", "qwerty", "abc123", "admin"}
	passwordLower := strings.ToLower(password)

	for _, pattern := range commonPatterns {
		if strings.Contains(passwordLower, pattern) {
			return errors.New("password contains a common pattern that is easily guessable")
		}
	}

	return nil
}
