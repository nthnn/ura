//go:build js && wasm
// +build js,wasm

package main

func main() {
	done := make(chan struct{}, 0)

	disableContextPopup()
	disableTextSelection()

	checkSessionKey()
	installOffcanvasListeners()

	loadInitialInformation()
	updateInformation()

	fixTabAnimations()
	installButtonActions()
	showActualContent()

	sessionValidationTicks()
	<-done
}
