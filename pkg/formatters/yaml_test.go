package formatters

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestYAMLFormatter(t *testing.T) {
	p := NewPluginYAMLFormatter()
	out := runFormat(t, p.Process, p.RegisterFlags, []byte("name: deen\nok: true\nitems:\n- one\n"))
	got := string(out)
	for _, want := range []string{"name: deen", "ok: true", "items:", "- one"} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatted YAML = %q, missing %q", got, want)
		}
	}
}

func TestYAMLToJSON(t *testing.T) {
	p := NewPluginYAMLFormatter()
	var out bytes.Buffer
	if err := p.Unprocess(bytes.NewReader([]byte("name: deen\nok: true\n")), &out, flag.NewFlagSet("yaml", flag.ContinueOnError)); err != nil {
		t.Fatalf("unprocess failed: %s", err)
	}
	if got, want := out.String(), "{\"name\":\"deen\",\"ok\":true}\n"; got != want {
		t.Fatalf("YAML to JSON = %q, want %q", got, want)
	}
}

func TestYAMLInvalid(t *testing.T) {
	p := NewPluginYAMLFormatter()
	if _, err := tryFormat(p.Process, p.RegisterFlags, []byte("name: [unterminated\n")); err == nil {
		t.Fatal("expected invalid YAML to fail")
	}
}
