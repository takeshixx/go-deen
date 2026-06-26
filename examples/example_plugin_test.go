package examples

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

func run(t *testing.T, fn types.TransformFunc, rf func(*flag.FlagSet), input []byte, args ...string) []byte {
	t.Helper()
	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	if rf != nil {
		rf(fs)
	}
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := fn(bytes.NewReader(input), &buf, fs); err != nil {
		t.Fatalf("transform failed: %s", err)
	}
	return buf.Bytes()
}

func TestPluginExampleRoundTrip(t *testing.T) {
	p := NewPluginExample()
	input := []byte("deen example plugin")
	encoded := run(t, p.Process, p.RegisterFlags, input)
	decoded := run(t, p.Unprocess, p.RegisterFlags, encoded)
	if !bytes.Equal(decoded, input) {
		t.Errorf("round-trip mismatch: got %q, want %q", decoded, input)
	}
}

func TestPluginStreamExample(t *testing.T) {
	p := NewPluginStreamExample()
	if p.Unprocess != nil {
		t.Error("one-way plugin should not define Unprocess")
	}
	got := run(t, p.Process, p.RegisterFlags, []byte("deenshatest"))
	if string(got) != "c324da7d32853ffaeb6577f624753c7f0f2842c0" {
		t.Errorf("unexpected SHA1: %s", got)
	}
}
