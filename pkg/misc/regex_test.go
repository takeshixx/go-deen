package misc

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"
)

func runMisc(t *testing.T, fn func(*flag.FlagSet), process func(io.Reader, io.Writer, *flag.FlagSet) error, input []byte, args ...string) []byte {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if fn != nil {
		fn(fs)
	}
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := process(bytes.NewReader(input), &out, fs); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func runMiscErr(fn func(*flag.FlagSet), process func(io.Reader, io.Writer, *flag.FlagSet) error, input []byte, args ...string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if fn != nil {
		fn(fs)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	var out bytes.Buffer
	return process(bytes.NewReader(input), &out, fs)
}

func TestRegexExtractGroup(t *testing.T) {
	p := NewPluginRegex()
	out := runMisc(t, p.RegisterFlags, p.Process, []byte("id=123 id=456"), "-re", `id=(\d+)`, "-group", "1")
	if string(out) != "123\n456" {
		t.Fatalf("regex extract = %q", string(out))
	}
}

func TestRegexReplace(t *testing.T) {
	p := NewPluginRegex()
	out := runMisc(t, p.RegisterFlags, p.Process, []byte("abc123"), "-re", `\d+`, "-replace", "###")
	if string(out) != "abc###" {
		t.Fatalf("regex replace = %q", string(out))
	}
}

func TestRegexMissingPattern(t *testing.T) {
	p := NewPluginRegex()
	fs := flag.NewFlagSet("regex", flag.ContinueOnError)
	p.RegisterFlags(fs)
	if err := p.Process(strings.NewReader("test"), &bytes.Buffer{}, fs); err == nil {
		t.Fatal("expected missing pattern to fail")
	}
}
