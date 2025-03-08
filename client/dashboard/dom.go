//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/base64"
	"errors"
	"syscall/js"

	"github.com/skip2/go-qrcode"
)

var document = js.Global().Get("document")

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

func fixTabAnimations() {
	tabLinks := document.Call("querySelectorAll", `a[data-bs-toggle="tab"]`)
	length := tabLinks.Length()

	for i := 0; i < length; i++ {
		tabLink := tabLinks.Index(i)
		callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			target := event.Get("target")

			targetSelector := target.Call("getAttribute", "data-bs-target")
			if targetSelector.Type() == js.TypeUndefined || targetSelector.String() == "" {
				targetSelector = target.Call("getAttribute", "href")
			}

			if targetSelector.Truthy() && targetSelector.String() != "" {
				pane := document.Call("querySelector", targetSelector.String())

				if pane.Truthy() {
					pane.Get("classList").Call("remove", "animate-slide")
					_ = pane.Get("offsetWidth").Int()

					pane.Get("classList").Call("add", "animate-slide")
				}
			}

			return nil
		})

		tabLink.Call("addEventListener", "shown.bs.tab", callback)
	}
}

func generateQRCode(id string, data string) error {
	png, err := qrcode.Encode(data, qrcode.Highest, 256)
	if err != nil {
		return errors.New("cannot create QR code")
	}

	base64Image := base64.StdEncoding.EncodeToString(png)
	dataURI := "data:image/png;base64," + base64Image

	document.Call("getElementById", id).Set("src", dataURI)
	return nil
}

func showActualContent() {
	document.Call(
		"getElementById",
		"card-security-code",
	).Set("value", getSessionKey("security_code"))

	js.Global().Call(
		"setTimeout",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadingContent := document.Call(
				"getElementById",
				"loading-content",
			)
			actualContent := document.Call(
				"getElementById",
				"actual-content",
			)
			navBar := document.Call(
				"getElementById",
				"navigation-bar",
			)

			loadingContent.Get("classList").Call("add", "d-none")
			actualContent.Get("classList").Call("remove", "d-none")
			actualContent.Get("classList").Call("add", "d-block")
			navBar.Get("classList").Call("remove", "d-none")

			return nil
		}),
		3000,
	)
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
}
