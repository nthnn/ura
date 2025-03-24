//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"
	"time"
)

func cashInEvent() {
	amount := getInputValue("cash-in-amount")
	if amount == "" {
		showError("cash-in-error", "Empty cash-in amount value.")
		return
	}

	status, _, content := sendPost(
		"/api/cashin",
		map[string]string{
			"amount": amount,
		},
		map[string]interface{}{
			"X-Session-Token": getSessionKey("session_token"),
			"X-Security-Code": getSessionKey("security_code"),
		},
	)

	if status != 200 {
		showError("cash-in-error", "Internal error occured.")
		return
	}

	var data map[string]string
	err := json.Unmarshal([]byte(content), &data)

	time.Sleep(1 * time.Second)
	hideLoading("cash-in")

	if err != nil || status != 200 {
		showError("cash-in-error", "Internal error occured.")
		return
	} else if value, exists := data["status"]; exists && value != "ok" {
		showError("cash-in-error", capitalizeFirst(data["message"]))
		return
	}

	if value, exists := data["transaction_id"]; exists {
		generateQRCode("cash-in-qr", value)

		mainContentClasses := document.Call(
			"getElementById",
			"main-cash-in-content",
		).Get("classList")
		qrContentClasses := document.Call(
			"getElementById",
			"qr-cash-in-content",
		).Get("classList")

		mainContentClasses.Call("remove", "d-block")
		mainContentClasses.Call("add", "d-none")

		qrContentClasses.Call("remove", "d-none")
		qrContentClasses.Call("add", "d-block")
	}
}

func withdrawEvent() {
	status, _, content := sendPost(
		"/api/withdraw",
		map[string]string{
			"amount": getInputValue("cash-out-amount"),
		},
		map[string]interface{}{
			"X-Session-Token": getSessionKey("session_token"),
			"X-Security-Code": getSessionKey("security_code"),
		},
	)

	if status != 200 {
		showError("cash-out-error", "Internal error occured.")
		return
	}

	var data map[string]string
	err := json.Unmarshal([]byte(content), &data)

	time.Sleep(1 * time.Second)
	hideLoading("cash-out")

	if err != nil || status != 200 {
		showError("cash-out-error", "Internal error occured.")
		return
	} else if value, exists := data["status"]; exists && value != "ok" {
		showError("cash-out-error", capitalizeFirst(data["message"]))
		return
	}

	if value, exists := data["transaction_id"]; exists {
		generateQRCode("cash-out-qr", value)

		mainContentClasses := document.Call(
			"getElementById",
			"main-cash-out-content",
		).Get("classList")
		qrContentClasses := document.Call(
			"getElementById",
			"qr-cash-out-content",
		).Get("classList")

		mainContentClasses.Call("remove", "d-block")
		mainContentClasses.Call("add", "d-none")

		qrContentClasses.Call("remove", "d-none")
		qrContentClasses.Call("add", "d-block")
	}
}

func installButtonActions() {
	securityCode := document.Call(
		"getElementById",
		"card-security-code",
	)
	showSecurityCodeButton := document.Call(
		"getElementById",
		"eye-show",
	)
	hideSecurityCodeButton := document.Call(
		"getElementById",
		"eye-hide",
	)

	cashInButton := document.Call(
		"getElementById",
		"cash-in-btn",
	)
	cashOutButton := document.Call(
		"getElementById",
		"cash-out-btn",
	)

	if securityCode.IsNull() || securityCode.IsUndefined() ||
		showSecurityCodeButton.IsNull() || showSecurityCodeButton.IsUndefined() ||
		hideSecurityCodeButton.IsNull() || hideSecurityCodeButton.IsUndefined() ||
		cashInButton.IsNull() || cashInButton.IsUndefined() ||
		cashOutButton.IsNull() || cashOutButton.IsUndefined() {
		return
	}

	disableSelect := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		args[0].Call("preventDefault")
		return nil
	})

	showSecurityCodeButton.Get("style").Set("userSelect", "none")
	showSecurityCodeButton.Call(
		"addEventListener",
		"selectstart",
		disableSelect,
	)
	showSecurityCodeButton.Call(
		"addEventListener",
		"mousedown",
		disableSelect,
	)
	showSecurityCodeButton.Call(
		"addEventListener",
		"click",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			securityCode.Call(
				"setAttribute",
				"type",
				"text",
			)

			showSecurityCodeButton.Get("classList").Call("remove", "d-block")
			showSecurityCodeButton.Get("classList").Call("add", "d-none")

			hideSecurityCodeButton.Get("classList").Call("remove", "d-none")
			hideSecurityCodeButton.Get("classList").Call("add", "d-block")

			return nil
		}),
	)

	hideSecurityCodeButton.Get("style").Set("userSelect", "none")
	hideSecurityCodeButton.Call(
		"addEventListener",
		"selectstart",
		disableSelect,
	)
	hideSecurityCodeButton.Call(
		"addEventListener",
		"mousedown",
		disableSelect,
	)
	hideSecurityCodeButton.Call(
		"addEventListener",
		"click",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			securityCode.Call(
				"setAttribute",
				"type",
				"password",
			)

			hideSecurityCodeButton.Get("classList").Call("remove", "d-block")
			hideSecurityCodeButton.Get("classList").Call("add", "d-none")

			showSecurityCodeButton.Get("classList").Call("remove", "d-none")
			showSecurityCodeButton.Get("classList").Call("add", "d-block")

			return nil
		}),
	)

	cashInButton.Call(
		"addEventListener",
		"click",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			showLoading("cash-in")
			go cashInEvent()

			return nil
		}),
	)

	cashOutButton.Call(
		"addEventListener",
		"click",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			showLoading("cash-out")
			go withdrawEvent()

			return nil
		}),
	)
}
