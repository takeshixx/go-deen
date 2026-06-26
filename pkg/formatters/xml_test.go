package formatters

import (
	"bytes"
	"strings"
	"testing"
)

func TestPluginXMLProcess(t *testing.T) {
	p := NewPluginXMLFormatter()
	input := []byte(`<root><a>1</a><b>2</b></root>`)
	got := runFormat(t, p.Process, p.RegisterFlags, input)
	if !strings.Contains(string(got), "\n    <a>1</a>") {
		t.Errorf("xml prettify did not indent:\n%s", got)
	}
}

func TestPluginXMLUnprocess(t *testing.T) {
	p := NewPluginXMLFormatter()
	pretty := []byte("<root>\n    <a>1</a>\n    <b>2</b>\n</root>")
	got := runFormat(t, p.Unprocess, p.RegisterFlags, pretty)
	if !bytes.Equal(got, []byte("<root><a>1</a><b>2</b></root>")) {
		t.Errorf("xml minify wrong: %q", got)
	}
}

func TestPluginXMLRoundTrip(t *testing.T) {
	p := NewPluginXMLFormatter()
	input := []byte(`<note><to>a</to><from>b</from><body>hi</body></note>`)
	pretty := runFormat(t, p.Process, p.RegisterFlags, input)
	mini := runFormat(t, p.Unprocess, p.RegisterFlags, pretty)
	if !bytes.Equal(mini, input) {
		t.Errorf("xml round-trip mismatch:\n got %q\nwant %q", mini, input)
	}
}

func TestPluginJSON2XML(t *testing.T) {
	p := NewPluginJSON2XML()
	json := []byte(`{"note":{"to":"a","from":"b"}}`)
	xml := runFormat(t, p.Process, p.RegisterFlags, json)
	for _, want := range []string{"<note>", "<to>a</to>", "<from>b</from>"} {
		if !strings.Contains(string(xml), want) {
			t.Errorf("json2xml missing %q in:\n%s", want, xml)
		}
	}
	// Round-trip XML back to JSON via Unprocess.
	back := runFormat(t, p.Unprocess, p.RegisterFlags, xml)
	if !strings.Contains(string(back), `"to": "a"`) || !strings.Contains(string(back), `"from": "b"`) {
		t.Errorf("xml2json round-trip lost data:\n%s", back)
	}
}

func TestPluginJSON2XMLBadInput(t *testing.T) {
	p := NewPluginJSON2XML()
	if _, err := tryFormat(p.Process, p.RegisterFlags, []byte("not json")); err == nil {
		t.Error("expected an error for invalid JSON input")
	}
}
