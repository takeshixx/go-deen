// +build gui

package main

import (
	"log"

	"github.com/takeshixx/deen/internal/gui"
)

func main() {
	// Spawn the GUI
	dg, err := gui.NewDeenGUI()
	if err != nil {
		log.Fatal(err)
	}
	dg.Run()
	return
}
