//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
	"unicode"
	"unicode/utf8"
)

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func redirectTo(url string) {
	js.Global().Get("window").Get("location").Set(
		"href",
		url,
	)
}
