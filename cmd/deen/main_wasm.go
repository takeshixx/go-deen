//go:build js && wasm

// WebAssembly entry point: launches the browser UI.
package main

import "github.com/takeshixx/deen/internal/webui"

func main() {
	webui.Run()
}
