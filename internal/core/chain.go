package core

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/pkg/helpers"
)

func runChain() int {
	fs := flag.NewFlagSet("chain", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of chain:\n\n")
		fmt.Fprintf(os.Stderr, "Run a saved deen Web/GUI transform chain.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  deen chain saved.json\n")
		fmt.Fprintf(os.Stderr, "  deen chain -file saved.json\n")
		fmt.Fprintf(os.Stderr, "  printf data | deen chain -stdin saved.json\n\n")
		fs.PrintDefaults()
	}
	chainFile := fs.String("file", "", "saved chain JSON file")
	inputFile := fs.String("input-file", "", "override chain source with this input file")
	stdinInput := fs.Bool("stdin", false, "override chain source with stdin")
	newline := fs.Bool("N", false, "append a trailing newline to the output")
	fs.Parse(helpers.RemoveBeforeSubcommand(os.Args, "chain"))

	args := fs.Args()
	path := *chainFile
	if path == "" && len(args) > 0 {
		path = args[0]
		args = args[1:]
	}
	if path == "" {
		fmt.Fprintln(os.Stderr, "deen: chain: missing chain file")
		return 2
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "deen: chain: failed to read chain file: %s\n", err)
		return 1
	}
	pipe := pipeline.New()
	if err := pipe.ImportJSON(data); err != nil {
		fmt.Fprintf(os.Stderr, "deen: chain: failed to import chain: %s\n", err)
		return 1
	}

	if *stdinInput || *inputFile != "" || len(args) > 0 {
		r, cleanup, err := selectChainInput(*inputFile, *stdinInput, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "deen: chain:", err)
			return 1
		}
		defer cleanup()
		input, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "deen: chain: failed to read input: %s\n", err)
			return 1
		}
		pipe.SetSourceOwned(input)
	}

	out := bufio.NewWriter(os.Stdout)
	if _, err := out.Write(pipe.Result()); err != nil {
		fmt.Fprintln(os.Stderr, "deen:", err)
		return 1
	}
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

	if step, err := firstChainError(pipe); err != nil {
		fmt.Fprintf(os.Stderr, "deen: chain: step %d (%s): %s\n", step+1, pipe.Steps()[step].Plugin, err)
		return 1
	}
	return 0
}

func selectChainInput(inputFile string, stdinInput bool, args []string) (io.Reader, func(), error) {
	noop := func() {}
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return nil, noop, fmt.Errorf("failed to open input file: %w", err)
		}
		return f, func() { f.Close() }, nil
	}
	if len(args) > 0 {
		return strings.NewReader(strings.Join(args, " ")), noop, nil
	}
	if stdinInput {
		return os.Stdin, noop, nil
	}
	return strings.NewReader(""), noop, nil
}

func firstChainError(pipe *pipeline.Pipeline) (int, error) {
	for i := range pipe.Steps() {
		if err := pipe.Err(i); err != nil {
			return i, err
		}
	}
	return -1, nil
}
