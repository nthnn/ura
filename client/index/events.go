//go:build js && wasm
// +build js,wasm

package main

import "syscall/js"

func loginEvent(this js.Value, args []js.Value) interface{} {
	username := getInputValue("login-username")
	password := getInputValue("login-password")

	hideError("login-error")

	if username == "" {
		showError("login-error", "Username cannot be empty.")
		return nil
	} else if !validateUsername(username) {
		showError("login-error", "Invalid username string.")
		return nil
	}

	go login(username, password)
	return nil
}

func signupEvent(this js.Value, args []js.Value) interface{} {
	username := getInputValue("signup-username")
	email := getInputValue("signup-email")
	password := getInputValue("signup-password")
	passwordConfirmation := getInputValue("signup-password-confirm")

	hideError("signup-error")
	hideError("signup-success")

	if username == "" {
		showError("signup-error", "Username cannot be empty.")
		return nil
	} else if !validateUsername(username) {
		showError("signup-error", "Invalid username string.")
		return nil
	}

	if email == "" {
		showError("signup-error", "Email address cannot be empty.")
		return nil
	}

	if !validateEmail(email) {
		showError("signup-error", "Invalid email address string format.")
		return nil
	}

	if err := validatePassword(password); err != nil {
		showError("signup-error", capitalizeFirst(err.Error()))
		return nil
	}

	if password != passwordConfirmation {
		showError("signup-error", "Password and confirmation did not matched.")
		return nil
	}

	go signup(username, email, password)
	return nil
}

func installOffcanvasListeners() {
	loginOffcanvas := document.Call(
		"getElementById",
		"login-offcanvas",
	)
	loginOffcanvasCloseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		document.Call(
			"getElementById",
			"login-username",
		).Set("value", "")

		document.Call(
			"getElementById",
			"login-password",
		).Set("value", "")
		return nil
	})

	signupOffcanvas := document.Call(
		"getElementById",
		"signup-offcanvas",
	)
	signupOffcanvasCloseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		document.Call(
			"getElementById",
			"signup-username",
		).Set("value", "")

		document.Call(
			"getElementById",
			"signup-email",
		).Set("value", "")

		document.Call(
			"getElementById",
			"signup-password",
		).Set("value", "")

		document.Call(
			"getElementById",
			"signup-password-confirm",
		).Set("value", "")
		return nil
	})

	loginOffcanvas.Call(
		"addEventListener",
		"hidden.bs.offcanvas",
		loginOffcanvasCloseCallback,
	)
	signupOffcanvas.Call(
		"addEventListener",
		"hidden.bs.offcanvas",
		signupOffcanvasCloseCallback,
	)
}
