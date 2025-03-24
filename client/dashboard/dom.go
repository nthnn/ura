//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html"
	"image/color"
	"syscall/js"

	"github.com/skip2/go-qrcode"
)

var document = js.Global().Get("document")

func setInputValue(id string, value string) {
	element := document.Call("getElementById", id)
	if !element.IsUndefined() && !element.IsNull() {
		element.Set("value", value)
	}
}

func showError(errorId string, message string) {
	errorText := js.Global().Get("document").Call(
		"getElementById",
		errorId,
	)

	if errorText.IsUndefined() || errorText.IsNull() {
		return
	}

	errorText.Get("classList").Call("remove", "d-none")
	errorText.Get("classList").Call("add", "d-block")
	errorText.Set("innerHTML", html.EscapeString(message))

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

	if errorText.IsUndefined() || errorText.IsNull() {
		return
	}

	errorText.Get("classList").Call("remove", "d-block")
	errorText.Get("classList").Call("add", "d-none")
	errorText.Set("innerHTML", "")
}

func showLoading(name string) {
	text := js.Global().Get("document").Call(
		"getElementById",
		name+"-text",
	)

	if text.IsUndefined() || text.IsNull() {
		return
	}

	loading := js.Global().Get("document").Call(
		"getElementById",
		name+"-loading",
	)

	if loading.IsUndefined() || loading.IsNull() {
		return
	}

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

	if text.IsUndefined() || text.IsNull() {
		return
	}

	loading := js.Global().Get("document").Call(
		"getElementById",
		name+"-loading",
	)

	if loading.IsUndefined() || loading.IsNull() {
		return
	}

	text.Get("classList").Call("remove", "d-none")
	text.Get("classList").Call("add", "d-block")

	loading.Get("classList").Call("remove", "d-block")
	loading.Get("classList").Call("add", "d-none")
}

func getInputValue(id string) string {
	element := document.Call("getElementById", id)
	if element.IsUndefined() || element.IsNull() {
		return ""
	}

	return element.Get("value").String()
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
	if style.IsUndefined() || style.IsNull() {
		return
	}

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
	png, err := qrcode.New(data, qrcode.Highest)
	errorValue := errors.New("cannot create QR code")

	if err != nil {
		return errorValue
	}

	png.ForegroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	png.BackgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	png.DisableBorder = true

	var buf bytes.Buffer
	err = png.Write(256, &buf)
	if err != nil {
		return errorValue
	}

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := "data:image/png;base64," + base64Img

	img := document.Call("getElementById", id)
	if img.IsNull() || img.IsUndefined() {
		return errors.New("cannot render image to DOM")
	}

	img.Set("src", dataURI)
	return nil
}

func showActualContent() {
	setInputValue(
		"card-security-code",
		getSessionKey("security_code"),
	)

	js.Global().Call(
		"setTimeout",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			loadingContent := document.Call(
				"getElementById",
				"loading-content",
			)

			if loadingContent.IsNull() || loadingContent.IsUndefined() {
				return nil
			}

			actualContent := document.Call(
				"getElementById",
				"actual-content",
			)

			if actualContent.IsNull() || actualContent.IsUndefined() {
				return nil
			}

			navBar := document.Call(
				"getElementById",
				"navigation-bar",
			)

			if navBar.IsNull() || navBar.IsUndefined() {
				return nil
			}

			loadingContent.Get("classList").Call("add", "d-none")
			actualContent.Get("classList").Call("remove", "d-none")
			actualContent.Get("classList").Call("add", "d-block")

			navBar.Get("classList").Call("remove", "d-none")
			return nil
		}),
		3000,
	)
}

func installOffcanvasListeners() {
	cashInOffcanvas := document.Call(
		"getElementById",
		"cash-in-offcanvas",
	)
	if !cashInOffcanvas.IsNull() && !cashInOffcanvas.IsUndefined() {
		cashInOffcanvas.Call(
			"addEventListener",
			"hidden.bs.offcanvas",
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				cashInAmount := document.Call(
					"getElementById",
					"cash-in-amount",
				)

				mainContentClasses := document.Call(
					"getElementById",
					"main-cash-in-content",
				).Get("classList")

				qrContentClasses := document.Call(
					"getElementById",
					"qr-cash-in-content",
				).Get("classList")

				qrCashInImage := document.Call(
					"getElementById",
					"cash-in-qr",
				)

				if cashInAmount.IsNull() || cashInAmount.IsUndefined() ||
					mainContentClasses.IsNull() || mainContentClasses.IsUndefined() ||
					qrContentClasses.IsNull() || qrContentClasses.IsUndefined() ||
					qrCashInImage.IsNull() || qrCashInImage.IsUndefined() {
					return nil
				}

				mainContentClasses.Call("remove", "d-none")
				mainContentClasses.Call("add", "d-block")

				qrContentClasses.Call("remove", "d-block")
				qrContentClasses.Call("add", "d-none")

				qrCashInImage.Set("src", "")
				cashInAmount.Set("value", "")

				return nil
			}),
		)
	}

	cashOutOffcanvas := document.Call(
		"getElementById",
		"cash-out-offcanvas",
	)
	if !cashOutOffcanvas.IsNull() && !cashOutOffcanvas.IsUndefined() {
		cashOutOffcanvas.Call(
			"addEventListener",
			"hidden.bs.offcanvas",
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				cashOutAmount := document.Call(
					"getElementById",
					"cash-out-amount",
				)

				mainContentClasses := document.Call(
					"getElementById",
					"main-cash-out-content",
				).Get("classList")

				qrContentClasses := document.Call(
					"getElementById",
					"qr-cash-out-content",
				).Get("classList")

				qrCashInImage := document.Call(
					"getElementById",
					"cash-out-qr",
				)

				if cashOutAmount.IsNull() || cashOutAmount.IsUndefined() ||
					mainContentClasses.IsNull() || mainContentClasses.IsUndefined() ||
					qrContentClasses.IsNull() || qrContentClasses.IsUndefined() ||
					qrCashInImage.IsNull() || qrCashInImage.IsUndefined() {
					return nil
				}

				mainContentClasses.Call("remove", "d-none")
				mainContentClasses.Call("add", "d-block")

				qrContentClasses.Call("remove", "d-block")
				qrContentClasses.Call("add", "d-none")

				qrCashInImage.Set("src", "")
				cashOutAmount.Set("value", "")

				return nil
			}),
		)
	}
}
