// +build js,wasm

package main

import (
	"github.com/takeshixx/deen/internal/web"
)

func main() {
	// Spawn web interface
	dw := web.NewDeenWeb()
	dw.Run()
}
