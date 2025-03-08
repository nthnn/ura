//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"
)

func redirectTo(url string) {
	js.Global().Get("window").Get("location").Set(
		"href",
		url,
	)
}

func numberWithCommas(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	parts := strings.Split(s, ".")

	intPart := parts[0]
	decPart := parts[1]

	var result strings.Builder
	n := len(intPart)

	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}

		result.WriteRune(c)
	}

	return result.String() + "." + decPart
}
