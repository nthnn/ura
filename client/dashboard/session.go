//go:build js && wasm
// +build js,wasm

package main

import "syscall/js"

func setSessionKey(key, value string) {
	js.Global().Get("sessionStorage").Call("setItem", key, value)
}

func getSessionKey(key string) string {
	val := js.Global().Get("sessionStorage").Call("getItem", key)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}

	return val.String()
}

func hasSessionKey(key string) bool {
	val := js.Global().Get("sessionStorage").Call("getItem", key)
	return !val.IsNull() && !val.IsUndefined()
}

func removeSessionKey(key string) {
	js.Global().Get("sessionStorage").Call("removeItem", key)
}
