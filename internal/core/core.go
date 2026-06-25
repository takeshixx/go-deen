package core

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/takeshixx/deen/internal/plugins"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var version string
var branch string
var printPluginsPtr *bool
var printPluginsJSONPtr *bool
var versionPtr *bool
var filePtr *string

// ParseFlags parses the global (pre-subcommand) flags and handles the
// informational ones (-l, -lj, -version) that exit immediately.
func ParseFlags() {
	printPluginsPtr = flag.Bool("l", false, "list available plugins")
	printPluginsJSONPtr = flag.Bool("lj", false, "list available plugins in JSON format")
	versionPtr = flag.Bool("version", false, "print version")
	filePtr = flag.String("file", "", "read input from file")
	flag.Parse()

	switch {
	case *printPluginsPtr:
		plugins.PrintAvailable(false)
		os.Exit(0)
	case *printPluginsJSONPtr:
		plugins.PrintAvailable(true)
		os.Exit(0)
	case *versionPtr:
		fmt.Print(version)
		if branch != "" {
			fmt.Printf("-%s", branch)
		}
		fmt.Print("\n")
		os.Exit(0)
	}
}

// RunCLI dispatches the requested plugin and returns a process exit code.
func RunCLI() int {
	cmd := flag.Arg(0)
	plugin, unprocess, ok := plugins.Resolve(cmd)
	if !ok {
		fmt.Fprintf(os.Stderr, "deen: invalid command: %q (use -l to list plugins)\n", cmd)
		return 2
	}

	// New unified contract.
	if plugin.Process != nil {
		return runUnified(plugin, unprocess, cmd)
	}

	// Legacy contract (plugins not yet ported to Process/Unprocess).
	return runLegacy(plugin, unprocess, cmd)
}

// runUnified drives a plugin implementing the Process/Unprocess contract.
func runUnified(plugin *types.DeenPlugin, unprocess bool, cmd string) int {
	transform := plugin.Process
	if unprocess {
		if plugin.Unprocess == nil {
			fmt.Fprintf(os.Stderr, "deen: %q does not support decoding\n", plugin.Name)
			return 2
		}
		transform = plugin.Unprocess
	}

	fs := flag.NewFlagSet(plugin.Name, flag.ExitOnError)
	fs.Usage = usageFor(fs, plugin)
	newline := fs.Bool("N", false, "append a trailing newline to the output")
	fileFlag := fs.String("file", "", "read input from file")
	if plugin.RegisterFlags != nil {
		plugin.RegisterFlags(fs)
	}
	fs.Parse(helpers.RemoveBeforeSubcommand(os.Args, cmd))

	reader, cleanup, err := selectInput(*fileFlag, fs.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "deen:", err)
		return 1
	}
	defer cleanup()

	out := bufio.NewWriter(os.Stdout)
	if err := transform(reader, out, fs); err != nil {
		fmt.Fprintf(os.Stderr, "deen: %s: %s\n", plugin.Name, err)
		return 1
	}
	if *newline {
		if _, err := out.WriteString("\n"); err != nil {
			fmt.Fprintln(os.Stderr, "deen:", err)
			return 1
		}
	}
	if err := out.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "deen:", err)
		return 1
	}
	return 0
}

// selectInput decides where input is read from: an explicit file (the global
// -file flag or a plugin-level -file flag), the remaining CLI arguments, or
// stdin. The returned cleanup closes the file when one was opened.
func selectInput(pluginFile string, args []string) (io.Reader, func(), error) {
	noop := func() {}
	path := *filePtr
	if path == "" {
		path = pluginFile
	}
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, noop, fmt.Errorf("failed to open file: %w", err)
		}
		return f, func() { f.Close() }, nil
	}
	if len(args) > 0 {
		return strings.NewReader(strings.Join(args, " ")), noop, nil
	}
	return os.Stdin, noop, nil
}

func usageFor(fs *flag.FlagSet, plugin *types.DeenPlugin) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", plugin.Name)
		if plugin.Description != "" {
			fmt.Fprintf(os.Stderr, "%s\n\n", plugin.Description)
		}
		fs.PrintDefaults()
	}
}

// runLegacy drives plugins still using the deprecated DeenTask/stream funcs.
// TODO: remove once every plugin implements the Process/Unprocess contract.
func runLegacy(plugin *types.DeenPlugin, unprocess bool, cmd string) int {
	plugin.Unprocess_ = unprocess

	var pluginParser *flag.FlagSet
	if plugin.AddDefaultCliFunc != nil {
		pluginParser = helpers.DefaultFlagSet()
		pluginParser = plugin.AddDefaultCliFunc(plugin, pluginParser, helpers.RemoveBeforeSubcommand(os.Args, cmd))
	}

	noNewLine := false
	if pluginParser != nil && pluginParser.Lookup("n") != nil && helpers.IsBoolFlag(pluginParser, "n") {
		noNewLine = true
	}

	task := types.NewDeenTask(os.Stdout)
	task.Command = plugin.Command

	// Decide where we read from: a file, data from argv or stdin.
	if *filePtr != "" {
		mainFile, err := os.Open(*filePtr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "deen: failed to open file:", err)
			return 1
		}
		defer mainFile.Close()
		task.Reader = mainFile
	} else if pluginParser != nil {
		pluginFilePtr := pluginParser.Lookup("file")
		if pluginFilePtr != nil && pluginFilePtr.Value.String() != "" {
			pluginFile, err := os.Open(pluginFilePtr.Value.String())
			if err != nil {
				fmt.Fprintln(os.Stderr, "deen: failed to open file:", err)
				return 1
			}
			defer pluginFile.Close()
			task.Reader = pluginFile
		} else if pluginParser.NArg() > 0 {
			task.Reader = strings.NewReader(strings.Join(pluginParser.Args(), " "))
		} else {
			task.Reader = os.Stdin
		}
	} else {
		if flag.NArg() > 1 {
			task.Reader = strings.NewReader(strings.Join(flag.Args()[1:], " "))
		} else {
			task.Reader = os.Stdin
		}
	}

	if plugin.ProcessDeenTaskFunc != nil {
		if plugin.ProcessDeenTaskWithFlags != nil || plugin.UnprocessDeenTaskWithFlags != nil {
			if unprocess {
				plugin.UnprocessDeenTaskWithFlags(pluginParser, task)
			} else {
				plugin.ProcessDeenTaskWithFlags(pluginParser, task)
			}
		} else {
			if unprocess {
				plugin.UnprocessDeenTaskFunc(task)
			} else {
				plugin.ProcessDeenTaskFunc(task)
			}
		}

		select {
		case err := <-task.ErrChan:
			fmt.Fprintf(os.Stderr, "deen: %s: %s\n", plugin.Name, err)
			return 1
		case <-task.DoneChan:
		}

		if !noNewLine {
			if _, err := io.WriteString(os.Stdout, "\n"); err != nil {
				fmt.Fprintln(os.Stderr, "deen:", err)
				return 1
			}
		}
		return 0
	}

	// Default stream implementation.
	var processedData []byte
	var err error
	if plugin.ProcessStreamWithCliFlagsFunc != nil || plugin.UnprocessStreamWithCliFlagsFunc != nil {
		if unprocess {
			processedData, err = plugin.UnprocessStreamWithCliFlagsFunc(pluginParser, task.Reader)
		} else {
			processedData, err = plugin.ProcessStreamWithCliFlagsFunc(pluginParser, task.Reader)
		}
	} else {
		if unprocess {
			processedData, err = plugin.UnprocessStreamFunc(task.Reader)
		} else {
			processedData, err = plugin.ProcessStreamFunc(task.Reader)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "deen: %s: %s\n", plugin.Name, err)
		return 1
	}

	os.Stdout.Write(processedData)
	if !noNewLine {
		io.WriteString(os.Stdout, "\n")
	}
	return 0
}
