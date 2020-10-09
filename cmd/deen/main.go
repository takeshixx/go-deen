package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/takeshixx/deen/internal/gui"
	"github.com/takeshixx/deen/internal/plugins"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var version = "v3.1.2-alpha"

func main() {
	noNewLinePtr := flag.Bool("n", false, "do not output the trailing newline")
	printPluginsPtr := flag.Bool("l", false, "list available plugins")
	printPluginsJSONPtr := flag.Bool("lj", false, "list available plugins in JSON format")
	versionPtr := flag.Bool("version", false, "print version")
	filePtr := flag.String("file", "", "read from file")
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
		plugin.Command = strings.TrimPrefix(cmd, ".")
		var pluginParser *flag.FlagSet
		if plugin.AddDefaultCliFunc != nil {
			pluginParser = helpers.DefaultFlagSet()
			pluginParser = plugin.AddDefaultCliFunc(plugin, pluginParser, helpers.RemoveBeforeSubcommand(os.Args, cmd))
		}

		if pluginParser != nil && pluginParser.Lookup("n") != nil && helpers.IsBoolFlag(pluginParser, "n") {
			tb := true
			noNewLinePtr = &tb
		}

		// Create a new task
		task := types.NewDeenTask(os.Stdout)
		task.Command = strings.TrimPrefix(cmd, ".")

		// Decide where we read from. Its either from a file,
		// data from argv or stdin.
		if *filePtr != "" {
			mainFile, err := os.Open(*filePtr)
			if err != nil {
				log.Fatalf("Failed to open file: %v", err)
			}
			task.Reader = mainFile
		} else if pluginParser != nil {
			pluginFilePtr := pluginParser.Lookup("file")
			if pluginFilePtr != nil && pluginFilePtr.Value.String() != "" {
				pluginFile, err := os.Open(pluginFilePtr.Value.String())
				if err != nil {
					log.Fatalf("Failed to open file: %v", err)
				}
				task.Reader = pluginFile
			} else if pluginParser.NArg() > 0 {
				inputData := strings.Join(pluginParser.Args(), " ")
				task.Reader = strings.NewReader(inputData)
			} else {
				task.Reader = os.Stdin
			}
		} else {
			if flag.NArg() > 1 {
				inputData := strings.Join(flag.Args()[1:], " ")
				task.Reader = strings.NewReader(inputData)
			} else {
				task.Reader = os.Stdin
			}
		}

		// TODO: this check will be removed as soon as the old
		// steam stubs have been ported to the new task-based
		// stubs.
		if plugin.ProcessDeenTaskFunc != nil {
			// PipeChanFunc prototype
			if plugin.ProcessDeenTaskWithFlags != nil || plugin.UnprocessDeenTaskWithFlags != nil {
				if plugin.Unprocess {
					plugin.UnprocessDeenTaskWithFlags(pluginParser, task)
				} else {
					plugin.ProcessDeenTaskWithFlags(pluginParser, task)
				}
			} else if plugin.ProcessDeenTaskFunc != nil || plugin.UnprocessDeenTaskFunc != nil {
				if plugin.Unprocess {
					plugin.UnprocessDeenTaskFunc(task)
				} else {
					plugin.ProcessDeenTaskFunc(task)
				}
			}

			select {
			case err := <-task.ErrChan:
				log.Fatalf("An error occured during plugin processing: %v\n", err)

			case <-task.DoneChan:
				//fmt.Fprintln(os.Stderr, "Plugin processing finished")
			}

			if !*noNewLinePtr {
				_, err = io.WriteString(os.Stdout, "\n")
				if err != nil {
					log.Fatal(err)
				}
			}

			return
		}

		// Default stream implementation (will be removed soon?)
		inputReader := task.Reader

		if plugin.ProcessStreamWithCliFlagsFunc != nil || plugin.UnprocessStreamWithCliFlagsFunc != nil {
			// Process plugins that actually use additional CLI flags
			if plugin.Unprocess {
				processedData, err = plugin.UnprocessStreamWithCliFlagsFunc(pluginParser, inputReader)
			} else {
				processedData, err = plugin.ProcessStreamWithCliFlagsFunc(pluginParser, inputReader)
			}
		} else {
			// Process plugins without additional CLI options
			if plugin.Unprocess {
				processedData, err = plugin.UnprocessStreamFunc(inputReader)
			} else {
				processedData, err = plugin.ProcessStreamFunc(inputReader)
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
		plugins.PrintAvailable(false)
	} else if *printPluginsJSONPtr {
		plugins.PrintAvailable(true)
	} else if *versionPtr {
		fmt.Println(version)
	} else {
		//flag.Usage()
		dg, err := gui.NewDeenGUI()
		if err != nil {
			log.Fatal(err)
		}
		dg.Run()
	}
}
