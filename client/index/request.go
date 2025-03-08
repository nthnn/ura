//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"syscall/js"
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
		close(resCh)

		return nil
	})
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errCh <- errors.New(args[0].String())
		close(errCh)

		return nil
	})
	fetchPromise.Call("then", thenFunc).Call("catch", catchFunc)

	var res js.Value
	select {
	case res = <-resCh:
	case err := <-errCh:
		thenFunc.Release()
		catchFunc.Release()

		return 0, "", err.Error()
	}

	thenFunc.Release()
	catchFunc.Release()

	status = res.Get("status").Int()
	contentType = res.Get("headers").Call("get", "Content-Type").String()

	textPromise := res.Call("text")
	textCh := make(chan js.Value)
	thenFuncText := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		textCh <- args[0]
		return nil
	})

	textPromise.Call("then", thenFuncText)
	textRes := <-textCh
	thenFuncText.Release()

	if textRes.Type() != js.TypeString {
		responseText = js.Global().Get("JSON").Call("stringify", textRes).String()
	} else {
		responseText = textRes.String()
	}

	return status, contentType, responseText
}
