package misc

import (
	"bytes"
	"flag"
	"regexp"
	"strings"
	"testing"
)

func TestUUIDGenerateV4(t *testing.T) {
	p := NewPluginUUID()
	fs := flag.NewFlagSet("uuid", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := fs.Parse([]string{"-gen"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.Process(strings.NewReader(""), &out, fs); err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(out.String()) {
		t.Fatalf("generated UUID has unexpected shape: %q", out.String())
	}
}

func TestUUIDTextToRawAndBack(t *testing.T) {
	p := NewPluginUUID()
	input := "550e8400-e29b-41d4-a716-446655440000"
	var raw bytes.Buffer
	if err := p.Unprocess(strings.NewReader(input), &raw, flag.NewFlagSet("uuid", flag.ContinueOnError)); err != nil {
		t.Fatal(err)
	}
	if raw.Len() != 16 {
		t.Fatalf("raw UUID length = %d, want 16", raw.Len())
	}
	var text bytes.Buffer
	if err := p.Process(bytes.NewReader(raw.Bytes()), &text, flag.NewFlagSet("uuid", flag.ContinueOnError)); err != nil {
		t.Fatal(err)
	}
	if text.String() != input {
		t.Fatalf("formatted UUID = %q, want %q", text.String(), input)
	}
}

func TestUUIDInfo(t *testing.T) {
	p := NewPluginUUID()
	fs := flag.NewFlagSet("uuid", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := fs.Parse([]string{"-info"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.Process(strings.NewReader("550e8400-e29b-41d4-a716-446655440000"), &out, fs); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"version: 4", "variant: RFC 4122"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("UUID info missing %q:\n%s", want, out.String())
		}
	}
}

func TestUUIDRejectsInvalid(t *testing.T) {
	p := NewPluginUUID()
	if err := p.Process(strings.NewReader("not-a-uuid"), &bytes.Buffer{}, flag.NewFlagSet("uuid", flag.ContinueOnError)); err == nil {
		t.Fatal("expected invalid UUID to fail")
	}
}
