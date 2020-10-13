// +build js,wasm

package main

import (
	"log"

	"github.com/takeshixx/deen/internal/web"
)

func main() {
	// Spawn web interface
	dw := web.NewDeenWeb()
	if err := dw.Run(); err != nil {
		log.Fatal(err)
	}
	return
}
