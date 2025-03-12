//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"time"
)

func checkSessionKey() {
	if hasSessionKey("session_token") && hasSessionKey("security_code") {
		go func() {
			status, _, content := sendPost(
				"/api/user/session",
				map[string]string{},
				map[string]interface{}{
					"X-Session-Token": getSessionKey("session_token"),
				},
			)

			var data map[string]string
			err := json.Unmarshal([]byte(content), &data)

			if err != nil || status != 200 {
				removeSessionKey("session_token")
				removeSessionKey("security_code")

				showActualContent()
				return
			}

			if value, exists := data["status"]; exists && value == "200" {
				redirectTo("/dashboard.html")
				return
			} else {
				removeSessionKey("session_token")
				removeSessionKey("security_code")
			}

			showActualContent()
		}()
	} else {
		showActualContent()
	}
}

func login(username, password string) {
	showLoading("login")

	hash := sha512.Sum512([]byte(password))
	status, _, content := sendPost(
		"/api/user/login",
		map[string]string{
			"username": username,
			"password": hex.EncodeToString(hash[:]),
		},
		map[string]interface{}{},
	)
	time.Sleep(2 * time.Second)

	var data map[string]string
	err := json.Unmarshal([]byte(content), &data)

	if err != nil || status != 200 {
		showError("login-error", "Internal error occured.")
		hideLoading("login")

		return
	} else if value, exists := data["status"]; exists && value != "200" {
		showError("login-error", capitalizeFirst(data["message"]))
		hideLoading("login")

		return
	} else {
		sessionToken, hasSessionToken := data["session_token"]
		securityCode, hasSecurityCode := data["security_code"]

		if hasSessionToken && hasSecurityCode {
			setSessionKey("session_token", sessionToken)
			setSessionKey("security_code", securityCode)

			redirectTo("/dashboard.html")
			return
		}

		showError("login-error", "Internal error occured.")
		hideLoading("login")

		return
	}
}

func signup(username, email, password string) {
	showLoading("signup")

	hash := sha512.Sum512([]byte(password))
	status, _, content := sendPost(
		"/api/user/create",
		map[string]string{
			"username": username,
			"email":    email,
			"password": hex.EncodeToString(hash[:]),
		},
		map[string]interface{}{},
	)
	time.Sleep(2 * time.Second)

	var data map[string]string
	err := json.Unmarshal([]byte(content), &data)

	if err != nil || status != 200 {
		showError("signup-error", "Internal error occured.")
		hideLoading("signup")

		return
	} else if value, exists := data["status"]; exists {
		if value == "200" {
			setInputValue("signup-username", "")
			setInputValue("signup-email", "")
			setInputValue("signup-password", "")
			setInputValue("signup-password-confirm", "")

			showError("signup-success", "Account created, you can now log-in.")
			hideLoading("signup")
			return
		} else {
			hideLoading("signup")

			if message, exists := data["message"]; exists {
				showError("signup-error", capitalizeFirst(message))
				return
			}

			showError("signup-error", "Internal error occured.")
			return
		}
	} else {
		showError("signup-error", "Failed to create new account. Please try again later.")
		hideLoading("signup")

		return
	}
}
