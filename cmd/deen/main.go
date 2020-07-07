package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/takeshixx/deen/internal/gui"
	"github.com/takeshixx/deen/internal/plugins"
)

var version = "v3.0-alpha"

func main() {
	noNewLinePtr := flag.Bool("n", false, "omit new line")
	printPluginsPtr := flag.Bool("l", false, "list available plugins")
	versionPtr := flag.Bool("version", false, "print version")
	flag.Parse()

	if flag.NArg() > 0 {
		var processedData []byte
		var err error
		cmd := flag.Arg(0)
		if !plugins.CmdAvailable(cmd) {
			fmt.Println("Inavlid cmd", cmd)
			return
		}
		plugin := plugins.GetForCmd(cmd)
		if strings.HasPrefix(cmd, ".") {
			plugin.Unprocess = true
		}
		var pluginParser *flag.FlagSet
		if plugin.AddCliOptionsFunc != nil {
			// Add CLI descriptions/options for plugins
			pluginParser = plugin.AddCliOptionsFunc(plugin, os.Args[2:])
		}
		if plugin.ProcessStreamWithCliFlagsFunc != nil || plugin.UnprocessStreamWithCliFlagsFunc != nil {
			// Process plugins that actually use additional CLI flags
			if pluginParser.NArg() > 0 {
				// Read data from CLI
				stringReader := strings.NewReader(pluginParser.Arg(0))
				if plugin.Unprocess {
					processedData, err = plugin.UnprocessStreamWithCliFlagsFunc(pluginParser, stringReader)
				} else {
					processedData, err = plugin.ProcessStreamWithCliFlagsFunc(pluginParser, stringReader)
				}
			} else {
				if plugin.Unprocess {
					processedData, err = plugin.UnprocessStreamWithCliFlagsFunc(pluginParser, os.Stdin)
				} else {
					// Read data from STDIN
					processedData, err = plugin.ProcessStreamWithCliFlagsFunc(pluginParser, os.Stdin)
				}
			}
		} else {
			// Process plugins without additional CLI options
			if flag.NArg() > 1 {
				stringReader := strings.NewReader(flag.Arg(1))
				if plugin.Unprocess {
					processedData, err = plugin.UnprocessStreamFunc(stringReader)
				} else {
					processedData, err = plugin.ProcessStreamFunc(stringReader)
				}
			} else {
				if plugin.Unprocess {
					processedData, err = plugin.UnprocessStreamFunc(os.Stdin)
				} else {
					processedData, err = plugin.ProcessStreamFunc(os.Stdin)
				}
			}
		}
		if err != nil {
			// TODO: better handle plugin errors?
			log.Fatal(err)
		}
		// Convert []byte to string
		outputData := fmt.Sprintf("%s", processedData)
		if *noNewLinePtr {
			fmt.Print(outputData)
		} else {
			fmt.Println(outputData)
		}
	} else if *printPluginsPtr {
		plugins.PrintAvailable()
	} else if *versionPtr {
		fmt.Println(version)
	} else {
		//flag.Usage()
		gui.RunGUI()
	}
}
