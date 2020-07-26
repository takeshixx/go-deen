package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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

		// pipeWriter is provided to plugins to write data to.
		// After a plugin has finished, processed data will be
		// piped from pipeReader to os.Stdout
		pipeReader, pipeWriter := io.Pipe()

		if plugin.ProcessPipeWithFlags != nil || plugin.UnprocessPipeWithFlags != nil {
			if pluginParser.NArg() > 0 {
				stringReader := strings.NewReader(pluginParser.Arg(0))
				if plugin.Unprocess {
					err = plugin.UnprocessPipeWithFlags(pluginParser, stringReader, pipeWriter)
				} else {
					err = plugin.ProcessPipeWithFlags(pluginParser, stringReader, pipeWriter)
				}
			} else {
				if plugin.Unprocess {
					err = plugin.UnprocessPipeWithFlags(pluginParser, os.Stdin, pipeWriter)
				} else {
					err = plugin.ProcessPipeWithFlags(pluginParser, os.Stdin, pipeWriter)
				}
			}

			io.Copy(os.Stdout, pipeReader)
			pipeReader.Close()
			if !*noNewLinePtr {
				_, err = io.WriteString(os.Stdout, "\n")
				if err != nil {
					log.Fatal(err)
				}
			}
			return
		} else if plugin.ProcessPipeFunc != nil || plugin.UnprocessPipeFunc != nil {
			if flag.NArg() > 1 {
				stringReader := strings.NewReader(flag.Arg(1))
				if plugin.Unprocess {
					err = plugin.UnprocessPipeFunc(stringReader, pipeWriter)
				} else {
					err = plugin.ProcessPipeFunc(stringReader, pipeWriter)
				}
			} else {
				if plugin.Unprocess {
					err = plugin.UnprocessPipeFunc(os.Stdin, pipeWriter)
				} else {
					err = plugin.ProcessPipeFunc(os.Stdin, pipeWriter)
				}
			}

			io.Copy(os.Stdout, pipeReader)
			pipeReader.Close()
			if !*noNewLinePtr {
				_, err = io.WriteString(os.Stdout, "\n")
				if err != nil {
					log.Fatal(err)
				}
			}
			return
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
		flag.Usage()
	}
}
