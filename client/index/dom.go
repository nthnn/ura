//go:build js && wasm
// +build js,wasm

package main

import "syscall/js"

var document = js.Global().Get("document")

func setEvent(id string, callback js.Func) {
	element := document.Call("getElementById", id)
	if element.IsUndefined() || element.IsNull() {
		return
	}

	element.Call("addEventListener", "click", callback)
}

func getInputValue(id string) string {
	element := document.Call("getElementById", id)
	if element.IsUndefined() || element.IsNull() {
		return ""
	}

	return element.Get("value").String()
}

func setInputValue(id string, value string) {
	element := document.Call("getElementById", id)
	if !element.IsUndefined() && !element.IsNull() {
		element.Set("value", value)
	}
}

func disableContextPopup() {
	contextMenuCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			args[0].Call("preventDefault")
		}

		return nil
	})

	keydownCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if e.Get("ctrlKey").Bool() && e.Get("shiftKey").Bool() && e.Get("keyCode").Int() == 73 {
			e.Call("preventDefault")
			return false
		}

		if e.Get("ctrlKey").Bool() && e.Get("keyCode").Int() == 67 {
			e.Call("preventDefault")
			return false
		}

		if e.Get("ctrlKey").Bool() && e.Get("keyCode").Int() == 82 {
			e.Call("preventDefault")
			return false
		}

		if e.Get("ctrlKey").Bool() && e.Get("keyCode").Int() == 85 {
			e.Call("preventDefault")
			return false
		}

		return nil
	})

	document.Call("addEventListener", "keydown", keydownCallback)
	document.Call("addEventListener", "contextmenu", contextMenuCallback)
}

func disableTextSelection() {
	style := document.Get("body").Get("style")

	style.Set("userSelect", "none")
	style.Set("-webkit-user-select", "none")
	style.Set("-moz-user-select", "none")
	style.Set("-ms-user-select", "none")
}

func showError(errorId string, message string) {
	errorText := js.Global().Get("document").Call(
		"getElementById",
		errorId,
	)

	errorText.Get("classList").Call("remove", "d-none")
	errorText.Get("classList").Call("add", "d-block")
	errorText.Set("innerHTML", message)

	var hideAfter2Secs js.Func
	hideAfter2Secs = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer hideAfter2Secs.Release()
		hideError(errorId)

		return nil
	})

	js.Global().Call(
		"setTimeout",
		hideAfter2Secs,
		2500,
	)
}

func hideError(errorId string) {
	errorText := js.Global().Get("document").Call(
		"getElementById",
		errorId,
	)

	errorText.Get("classList").Call("remove", "d-block")
	errorText.Get("classList").Call("add", "d-none")
	errorText.Set("innerHTML", "")
}

func showLoading(name string) {
	text := js.Global().Get("document").Call(
		"getElementById",
		name+"-text",
	)
	loading := js.Global().Get("document").Call(
		"getElementById",
		name+"-loading",
	)

	text.Get("classList").Call("remove", "d-block")
	text.Get("classList").Call("add", "d-none")

	loading.Get("classList").Call("remove", "d-none")
	loading.Get("classList").Call("add", "d-block")
}

func hideLoading(name string) {
	text := js.Global().Get("document").Call(
		"getElementById",
		name+"-text",
	)
	loading := js.Global().Get("document").Call(
		"getElementById",
		name+"-loading",
	)

	text.Get("classList").Call("remove", "d-none")
	text.Get("classList").Call("add", "d-block")

	loading.Get("classList").Call("remove", "d-block")
	loading.Get("classList").Call("add", "d-none")
}

func showActualContent() {
	js.Global().Call(
		"setTimeout",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadingContent := js.Global().Get("document").Call(
				"getElementById",
				"loading-content",
			)
			actualContent := js.Global().Get("document").Call(
				"getElementById",
				"actual-content",
			)

			loadingContent.Get("classList").Call("add", "d-none")
			actualContent.Get("classList").Call("remove", "d-none")
			actualContent.Get("classList").Call("add", "d-block")

			return nil
		}),
		3000,
	)
}
