//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"sync"
	"syscall/js"
	"time"
)

func checkSessionKey() {
	var wg sync.WaitGroup
	wg.Add(1)

	if hasSessionKey("session_token") && hasSessionKey("security_code") {
		go func() {
			defer wg.Done()

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

				redirectTo("/")
				return
			}

			if value, exists := data["status"]; !exists || value != "200" {
				removeSessionKey("session_token")
				removeSessionKey("security_code")

				redirectTo("/")
				return
			}
		}()
	} else {
		removeSessionKey("session_token")
		removeSessionKey("security_code")

		redirectTo("/")
	}

	wg.Wait()
}

func installOffcanvasListeners() {
	cashInOffcanvas := document.Call(
		"getElementById",
		"cash-in-offcanvas",
	)
	cashInOffcanvasCloseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		document.Call(
			"getElementById",
			"cash-in-amount",
		).Set("value", "")
		return nil
	})

	cashOutOffcanvas := document.Call(
		"getElementById",
		"cash-out-offcanvas",
	)
	cashOutOffcanvasCloseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		document.Call(
			"getElementById",
			"cash-out-amount",
		).Set("value", "")
		return nil
	})

	cashInOffcanvas.Call(
		"addEventListener",
		"hidden.bs.offcanvas",
		cashInOffcanvasCloseCallback,
	)
	cashOutOffcanvas.Call(
		"addEventListener",
		"hidden.bs.offcanvas",
		cashOutOffcanvasCloseCallback,
	)
}

func periodicSessionValidation() {
	if hasSessionKey("session_token") || hasSessionKey("security_code") {
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

				redirectTo("/")
				return
			}

			if value, exists := data["status"]; !exists || value != "200" {
				removeSessionKey("session_token")
				removeSessionKey("security_code")

				redirectTo("/")
				return
			}

			if value, exists := data["expired"]; exists && value == "true" {
				removeSessionKey("session_token")
				removeSessionKey("security_code")

				redirectTo("/")
				return
			}
		}()
	} else {
		removeSessionKey("session_token")
		removeSessionKey("security_code")

		redirectTo("/")
	}
}

func sessionValidationTicks() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			periodicSessionValidation()
		}
	}()
}
