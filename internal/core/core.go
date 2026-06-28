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

// Version returns the build version (set via -ldflags), or "dev" if unset.
func Version() string {
	if version == "" {
		return "dev"
	}
	return version
}

// Branch returns the build branch (set via -ldflags), or "" if unset.
func Branch() string { return branch }

var printPluginsPtr *bool
var printPluginsJSONPtr *bool
var versionPtr *bool
var filePtr *string
var newlinePtr *bool

// ParseFlags parses the global (pre-subcommand) flags and handles the
// informational ones (-l, -lj, -version) that exit immediately.
func ParseFlags() {
	printPluginsPtr = flag.Bool("l", false, "list available plugins")
	printPluginsJSONPtr = flag.Bool("lj", false, "list available plugins in JSON format")
	versionPtr = flag.Bool("version", false, "print version")
	filePtr = flag.String("file", "", "read input from file")
	newlinePtr = flag.Bool("N", false, "append a trailing newline to the output")
	flag.Usage = printCLIUsage
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

func printCLIUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, "deen - encode, decode, hash, compress and inspect data")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  deen [global flags] <plugin> [plugin flags] [input]")
	fmt.Fprintln(out, "  deen [global flags] .<plugin> [plugin flags] [input]")
	fmt.Fprintln(out, "  deen chain [chain flags] <chain.json> [input]")
	fmt.Fprintln(out, "  deen serve [serve flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Run a plugin by name to transform data. Prefix a reversible plugin with '.'")
	fmt.Fprintln(out, "to run its decode/unprocess direction, for example '.base64' or '.gzip'.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Input:")
	fmt.Fprintln(out, "  Input is read from remaining arguments, stdin, or -file.")
	fmt.Fprintln(out, "  Output is raw by default, without a trailing newline; pass -N for terminals.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Common commands:")
	fmt.Fprintln(out, "  deen -l                         list plugins by category")
	fmt.Fprintln(out, "  deen -lj                        list plugins as JSON")
	fmt.Fprintln(out, "  deen base64 test                encode text as Base64")
	fmt.Fprintln(out, "  deen .base64 dGVzdA==           decode Base64")
	fmt.Fprintln(out, "  deen chain saved.json           run a saved Web/GUI chain")
	fmt.Fprintln(out, "  printf secret | deen sha256     hash stdin")
	fmt.Fprintln(out, "  deen base64 -h                  show plugin-specific flags")
	fmt.Fprintln(out, "  deen serve --port 9090          serve the WebAssembly UI")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Global flags:")
	flag.PrintDefaults()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Use 'deen <plugin> -h' for plugin flags, 'deen chain -h' for saved chains,")
	fmt.Fprintln(out, "or 'deen serve -h' for web server flags.")
}

// RunCLI dispatches the requested plugin and returns a process exit code.
func RunCLI() int {
	cmd := flag.Arg(0)
	if cmd == "serve" {
		return runServe()
	}
	if cmd == "chain" {
		return runChain()
	}
	plugin, unprocess, ok := plugins.Resolve(cmd)
	if !ok {
		fmt.Fprintf(os.Stderr, "deen: invalid command: %q (use -l to list plugins)\n", cmd)
		return 2
	}
	return runPlugin(plugin, unprocess, cmd)
}

// runPlugin drives a plugin implementing the Process/Unprocess contract.
func runPlugin(plugin *types.DeenPlugin, unprocess bool, cmd string) int {
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
		if flushErr := out.Flush(); flushErr != nil {
			fmt.Fprintln(os.Stderr, "deen:", flushErr)
			return 1
		}
		fmt.Fprintf(os.Stderr, "deen: %s: %s\n", plugin.Name, err)
		return 1
	}
	// The newline may be requested either globally (before the plugin name)
	// or as a per-plugin flag (after it).
	if *newline || *newlinePtr {
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
