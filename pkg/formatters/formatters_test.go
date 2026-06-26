package formatters

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

// tryFormat runs a transform with a fresh flag set parsed from args.
func tryFormat(fn types.TransformFunc, registerFlags func(*flag.FlagSet), input []byte, args ...string) ([]byte, error) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if registerFlags != nil {
		registerFlags(fs)
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err := fn(bytes.NewReader(input), &buf, fs)
	return buf.Bytes(), err
}

// runFormat is tryFormat but fails the test on error.
func runFormat(t *testing.T, fn types.TransformFunc, registerFlags func(*flag.FlagSet), input []byte, args ...string) []byte {
	t.Helper()
	out, err := tryFormat(fn, registerFlags, input, args...)
	if err != nil {
		t.Fatalf("transform failed: %s", err)
	}
	return out
}
