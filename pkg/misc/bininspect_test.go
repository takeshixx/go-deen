package misc

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestBinInspectCurrentExecutable(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPluginBinInspect()
	var out bytes.Buffer
	if err := p.Process(bytes.NewReader(data), &out, nil); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"format:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("bininspect output missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "sections:") && !strings.Contains(text, "arches:") {
		t.Fatalf("bininspect output missing sections or arches:\n%s", text)
	}
}

func TestBinInspectRejectsUnsupportedInput(t *testing.T) {
	p := NewPluginBinInspect()
	if err := p.Process(strings.NewReader("not an executable"), &bytes.Buffer{}, nil); err == nil {
		t.Fatal("expected unsupported input to fail")
	}
}
