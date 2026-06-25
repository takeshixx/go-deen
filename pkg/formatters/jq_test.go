package formatters

import (
	"strings"
	"testing"
)

func TestPluginJQProcess(t *testing.T) {
	p := NewPluginJQFormatter()
	input := []byte(`{"name": "deen", "nested": {"value": 42}}`)
	got := runFormat(t, p.Process, p.RegisterFlags, input, "-q", ".nested.value", "-plain")
	if strings.TrimSpace(string(got)) != "42" {
		t.Errorf("jq query: got %q, want 42", got)
	}
}

func TestPluginJQNoQuery(t *testing.T) {
	p := NewPluginJQFormatter()
	if _, err := tryFormat(p.Process, p.RegisterFlags, []byte("{}")); err == nil {
		t.Error("expected an error when no query is provided")
	}
}
