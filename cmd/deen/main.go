//go:build !gui && !js

package main

import (
	"flag"
	"os"

	"github.com/takeshixx/deen/internal/core"
)

func main() {
	core.ParseFlags()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	os.Exit(core.RunCLI())
}
