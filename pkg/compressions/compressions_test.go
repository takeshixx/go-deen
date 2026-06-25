package compressions

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

var compTestData = []byte("deen compression test data: the quick brown fox jumps over the lazy dog 0123456789 0123456789")

// transform runs fn with a fresh flag set parsed from args.
func transform(fn types.TransformFunc, registerFlags func(*flag.FlagSet), input []byte, args ...string) ([]byte, error) {
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

// assertRoundTrip compresses input then decompresses it and checks it matches.
func assertRoundTrip(t *testing.T, p *types.DeenPlugin, args ...string) {
	t.Helper()
	compressed, err := transform(p.Process, p.RegisterFlags, compTestData, args...)
	if err != nil {
		t.Fatalf("%s compress failed: %s", p.Name, err)
	}
	out, err := transform(p.Unprocess, p.RegisterFlags, compressed, args...)
	if err != nil {
		t.Fatalf("%s decompress failed: %s", p.Name, err)
	}
	if !bytes.Equal(out, compTestData) {
		t.Errorf("%s round-trip mismatch", p.Name)
	}
}

// assertDecompressError checks that decompressing junk returns an error
// instead of panicking.
func assertDecompressError(t *testing.T, p *types.DeenPlugin) {
	t.Helper()
	if _, err := transform(p.Unprocess, p.RegisterFlags, []byte("this is definitely not compressed data")); err == nil {
		t.Errorf("%s: expected an error decompressing junk input", p.Name)
	}
}
