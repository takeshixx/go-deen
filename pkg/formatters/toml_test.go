package formatters

import (
	"strings"
	"testing"
)

func TestTOMLFormatter(t *testing.T) {
	p := NewPluginTOML()
	out := runFormat(t, p.Process, p.RegisterFlags, []byte("name = \"deen\"\n[meta]\nok = true\n"))
	got := string(out)
	for _, want := range []string{"name = \"deen\"", "[meta]", "ok = true"} {
		if !strings.Contains(got, want) {
			t.Fatalf("TOML output missing %q:\n%s", want, got)
		}
	}
}

func TestTOMLToJSON(t *testing.T) {
	p := NewPluginTOML()
	out := runFormat(t, p.Unprocess, p.RegisterFlags, []byte("name = \"deen\"\n"))
	if !strings.Contains(string(out), `"name": "deen"`) {
		t.Fatalf("TOML to JSON unexpected:\n%s", string(out))
	}
}
