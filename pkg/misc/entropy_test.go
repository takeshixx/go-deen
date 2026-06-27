package misc

import (
	"bytes"
	"strings"
	"testing"
)

func TestEntropyAnalyzer(t *testing.T) {
	p := NewPluginEntropy()
	var out bytes.Buffer
	if err := p.Process(strings.NewReader("aaaa"), &out, nil); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"bytes: 4", "unique: 1", "entropy: 0.0000 bits/byte", "0x61: 4 100.00%"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("entropy output missing %q:\n%s", want, out.String())
		}
	}
}

func TestEntropyEmptyInput(t *testing.T) {
	p := NewPluginEntropy()
	var out bytes.Buffer
	if err := p.Process(bytes.NewReader(nil), &out, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "bytes: 0") {
		t.Fatalf("empty entropy output unexpected:\n%s", out.String())
	}
}
