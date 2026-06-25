package codecs

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

// tryCodec runs a transform with a fresh flag set built from registerFlags.
func tryCodec(fn types.TransformFunc, registerFlags func(*flag.FlagSet), input []byte, args ...string) ([]byte, error) {
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

// runCodec is tryCodec but fails the test on error.
func runCodec(t *testing.T, fn types.TransformFunc, registerFlags func(*flag.FlagSet), input []byte, args ...string) []byte {
	t.Helper()
	out, err := tryCodec(fn, registerFlags, input, args...)
	if err != nil {
		t.Fatalf("codec transform failed: %s", err)
	}
	return out
}

// assertCodec checks a transform produces want.
func assertCodec(t *testing.T, p *types.DeenPlugin, fn types.TransformFunc, input, want []byte, args ...string) {
	t.Helper()
	if got := runCodec(t, fn, p.RegisterFlags, input, args...); !bytes.Equal(got, want) {
		t.Errorf("%s: got %q, want %q", p.Name, got, want)
	}
}
