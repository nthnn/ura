//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

var sessionStorage js.Value = js.Global().Get("sessionStorage")

func isSessionStorageSupported() bool {
	return !sessionStorage.IsNull() && !sessionStorage.IsUndefined()
}

func setSessionKey(key, value string) {
	if isSessionStorageSupported() {
		sessionStorage.Call("setItem", key, value)
	}
}

func getSessionKey(key string) string {
	if isSessionStorageSupported() {
		val := sessionStorage.Call("getItem", key)
		if val.IsNull() || val.IsUndefined() {
			return ""
		}

		return val.String()
	}

	return ""
}

func hasSessionKey(key string) bool {
	if isSessionStorageSupported() {
		val := sessionStorage.Call("getItem", key)
		return !val.IsNull() && !val.IsUndefined()
	}

	return false
}

func removeSessionKey(key string) {
	if isSessionStorageSupported() {
		sessionStorage.Call("removeItem", key)
	}
}
