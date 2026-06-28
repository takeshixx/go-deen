package arithmetic

import (
	"bytes"
	"flag"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

func runArithmetic(t *testing.T, p *types.DeenPlugin, fn types.TransformFunc, input []byte, args ...string) []byte {
	t.Helper()
	fs := flag.NewFlagSet(p.Name, flag.ContinueOnError)
	if p.RegisterFlags != nil {
		p.RegisterFlags(fs)
	}
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := fn(bytes.NewReader(input), &out, fs); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestXORRoundTrip(t *testing.T) {
	p := NewPluginXOR()
	input := []byte{0x00, 0x41, 0xff}
	encoded := runArithmetic(t, p, p.Process, input, "-value", "0x2a")
	decoded := runArithmetic(t, p, p.Unprocess, encoded, "-value", "0x2a")
	if !bytes.Equal(decoded, input) {
		t.Fatalf("decoded %#v, want %#v", decoded, input)
	}
}

func TestAddRoundTrip(t *testing.T) {
	p := NewPluginAdd()
	input := []byte{0x00, 0xfe, 0xff}
	encoded := runArithmetic(t, p, p.Process, input, "-value", "2")
	decoded := runArithmetic(t, p, p.Unprocess, encoded, "-value", "2")
	if !bytes.Equal(decoded, input) {
		t.Fatalf("decoded %#v, want %#v", decoded, input)
	}
}

func TestSubRoundTrip(t *testing.T) {
	p := NewPluginSub()
	input := []byte{0x00, 0x01, 0xff}
	encoded := runArithmetic(t, p, p.Process, input, "-value", "2")
	decoded := runArithmetic(t, p, p.Unprocess, encoded, "-value", "2")
	if !bytes.Equal(decoded, input) {
		t.Fatalf("decoded %#v, want %#v", decoded, input)
	}
}

func TestNotRoundTrip(t *testing.T) {
	p := NewPluginNot()
	input := []byte{0x00, 0x55, 0xff}
	encoded := runArithmetic(t, p, p.Process, input)
	decoded := runArithmetic(t, p, p.Unprocess, encoded)
	if !bytes.Equal(decoded, input) {
		t.Fatalf("decoded %#v, want %#v", decoded, input)
	}
}
