package misc

import (
	"bytes"
	"strings"
	"testing"
)

func TestMagicDetectsPNG(t *testing.T) {
	p := NewPluginMagic()
	var out bytes.Buffer
	input := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	if err := p.Process(bytes.NewReader(input), &out, nil); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type: PNG image", "mime: image/png"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("magic output missing %q:\n%s", want, out.String())
		}
	}
}

func TestMagicFallsBackToContentSniffing(t *testing.T) {
	p := NewPluginMagic()
	var out bytes.Buffer
	if err := p.Process(strings.NewReader("hello"), &out, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "mime: text/plain") {
		t.Fatalf("magic fallback unexpected:\n%s", out.String())
	}
}
