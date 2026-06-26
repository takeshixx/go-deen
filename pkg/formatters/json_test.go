package formatters

import (
	"bytes"
	"testing"
)

var jsonTestData = []byte(`{"version": "v1.0", "enable": true}`)
var jsonTestDataProcessed = []byte(`{
    "enable": true,
    "version": "v1.0"
}
`)
var jsonTestDataMinified = []byte(`{"enable":true,"version":"v1.0"}`)

func TestPluginJSONProcess(t *testing.T) {
	p := NewPluginJSONFormatter()
	got := runFormat(t, p.Process, p.RegisterFlags, jsonTestData, "-no-color")
	if !bytes.Equal(got, jsonTestDataProcessed) {
		t.Errorf("json prettify: got %q, want %q", got, jsonTestDataProcessed)
	}
}

func TestPluginJSONUnprocess(t *testing.T) {
	p := NewPluginJSONFormatter()
	got := runFormat(t, p.Unprocess, p.RegisterFlags, jsonTestDataProcessed)
	if !bytes.Equal(got, jsonTestDataMinified) {
		t.Errorf("json minify: got %q, want %q", got, jsonTestDataMinified)
	}
}
