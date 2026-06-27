package misc

import (
	"bytes"
	"strings"
	"testing"
)

func runASN1(t *testing.T, input []byte) string {
	t.Helper()
	p := NewPluginASN1()
	var out bytes.Buffer
	if err := p.Process(bytes.NewReader(input), &out, nil); err != nil {
		t.Fatalf("asn1 Process failed: %s", err)
	}
	return out.String()
}

func TestASN1DERTree(t *testing.T) {
	input := []byte{
		0x30, 0x0b,
		0x02, 0x01, 0x2a,
		0x06, 0x06, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d,
	}
	out := runASN1(t, input)
	for _, want := range []string{
		"SEQUENCE",
		"INTEGER = 42",
		"OBJECT IDENTIFIER = 1.2.840.113549",
		"@0002",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("asn1 output missing %q\n%s", want, out)
		}
	}
}

func TestASN1PEMInput(t *testing.T) {
	input := []byte("-----BEGIN TEST-----\nMAMCASo=\n-----END TEST-----\n")
	out := runASN1(t, input)
	if !strings.Contains(out, "INTEGER = 42") {
		t.Fatalf("asn1 failed to parse PEM input\n%s", out)
	}
}

func TestASN1RejectsInvalidDER(t *testing.T) {
	p := NewPluginASN1()
	if err := p.Process(bytes.NewReader([]byte{0x30, 0x10, 0x02}), &bytes.Buffer{}, nil); err == nil {
		t.Fatal("expected invalid DER to fail")
	}
}
