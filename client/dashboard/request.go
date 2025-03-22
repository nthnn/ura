//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"syscall/js"
	"time"
)

func sendPost(
	urlStr string,
	data map[string]string,
	headers map[string]interface{},
) (
	status int,
	contentType string,
	responseText string,
) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, "", "Invalid data form."
	}

	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	opts := js.ValueOf(map[string]interface{}{
		"method":  "POST",
		"body":    string(jsonBody),
		"headers": headers,
	})

	fetchPromise := js.Global().Call("fetch", urlStr, opts)
	resCh := make(chan js.Value)
	errCh := make(chan error)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resCh <- args[0]
		return nil
	})
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errCh <- errors.New(args[0].String())
		return nil
	})

	fetchPromise.Call(
		"then",
		thenFunc,
	).Call(
		"catch",
		catchFunc,
	)

	select {
	case res := <-resCh:
		thenFunc.Release()
		catchFunc.Release()

		status = res.Get("status").Int()
		contentType = res.Get("headers").Call(
			"get",
			"Content-Type",
		).String()

		textPromise := res.Call("text")
		textCh := make(chan js.Value)

		thenFuncText := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			textCh <- args[0]
			return nil
		})

		catchFuncText := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			errCh <- errors.New(args[0].String())
			return nil
		})

		textPromise.Call(
			"then",
			thenFuncText,
		).Call(
			"catch",
			catchFuncText,
		)

		select {
		case textRes := <-textCh:
			thenFuncText.Release()
			catchFuncText.Release()

			if textRes.Type() != js.TypeString {
				responseText = js.Global().Get("JSON").Call("stringify", textRes).String()
			} else {
				responseText = textRes.String()
			}

		case err := <-errCh:
			thenFuncText.Release()
			catchFuncText.Release()

			return 0, "", err.Error()

		case <-time.After(10 * time.Second):
			thenFuncText.Release()
			catchFuncText.Release()

			return 0, "", "Timeout waiting for text response."
		}

		return status, contentType, responseText

	case err := <-errCh:
		thenFunc.Release()
		catchFunc.Release()

		return 0, "", err.Error()

	case <-time.After(10 * time.Second):
		thenFunc.Release()
		catchFunc.Release()

		return 0, "", "Timeout waiting for fetch response."
	}
}
