package hashs

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

// tryHash runs a plugin's Process with the given args and returns its output.
func tryHash(p *types.DeenPlugin, input []byte, args ...string) ([]byte, error) {
	fs := flag.NewFlagSet(p.Name, flag.ContinueOnError)
	if p.RegisterFlags != nil {
		p.RegisterFlags(fs)
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := p.Process(bytes.NewReader(input), &buf, fs); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// runHash is tryHash but fails the test on error.
func runHash(t *testing.T, p *types.DeenPlugin, input []byte, args ...string) []byte {
	t.Helper()
	out, err := tryHash(p, input, args...)
	if err != nil {
		t.Fatalf("%s process failed: %s", p.Name, err)
	}
	return out
}

// assertHash checks the produced digest matches want.
func assertHash(t *testing.T, p *types.DeenPlugin, input []byte, want string, args ...string) {
	t.Helper()
	if got := string(runHash(t, p, input, args...)); got != want {
		t.Errorf("%s: got %q, want %q", p.Name, got, want)
	}
}
