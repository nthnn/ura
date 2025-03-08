//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

func main() {
	done := make(chan struct{}, 0)

	disableContextPopup()
	disableTextSelection()
	installOffcanvasListeners()
	checkSessionKey()

	loginCallback := js.FuncOf(loginEvent)
	defer loginCallback.Release()

	signupCallback := js.FuncOf(signupEvent)
	defer signupCallback.Release()

	setEvent("login-btn", loginCallback)
	setEvent("signup-btn", signupCallback)

	<-done
}
