//go:build js && wasm
// +build js,wasm

package main

import (
	"net/url"
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

func redirectTo(link string) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return
	}

	location := js.Global().Get("window").Get("location")
	if !location.IsNull() && !location.IsUndefined() {
		location.Set(
			"href",
			link,
		)
	}
}
