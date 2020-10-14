// +build gui,!js,!wasm

package main

import (
	"flag"
	"log"

	"github.com/takeshixx/deen/internal/core"
	"github.com/takeshixx/deen/internal/gui"
)

func main() {
	core.ParseFlags()
	if flag.NArg() > 0 {
		core.RunCLI()
	} else {
		// Spawn the GUI
		dg, err := gui.NewDeenGUI()
		if err != nil {
			log.Fatal(err)
		}
		dg.Run()
	}
}
