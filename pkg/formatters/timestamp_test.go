package formatters

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestTimestampProcessSeconds(t *testing.T) {
	p := NewPluginTimestamp()
	got := runTimestamp(t, p.Process, p.RegisterFlags, []byte("0"))
	if got != "1970-01-01T00:00:00Z" {
		t.Fatalf("got %q", got)
	}
}

func TestTimestampProcessAutoMilliseconds(t *testing.T) {
	p := NewPluginTimestamp()
	got := runTimestamp(t, p.Process, p.RegisterFlags, []byte("1700000000123"))
	if got != "2023-11-14T22:13:20.123Z" {
		t.Fatalf("got %q", got)
	}
}

func TestTimestampUnprocessSeconds(t *testing.T) {
	p := NewPluginTimestamp()
	got := runTimestamp(t, p.Unprocess, p.RegisterFlags, []byte("1970-01-01T00:00:00Z"))
	if got != "0" {
		t.Fatalf("got %q", got)
	}
}

func TestTimestampUnprocessMilliseconds(t *testing.T) {
	p := NewPluginTimestamp()
	got := runTimestamp(t, p.Unprocess, p.RegisterFlags, []byte("2023-11-14T22:13:20.123Z"), "-unit", "ms")
	if got != "1700000000123" {
		t.Fatalf("got %q", got)
	}
}

func TestTimestampCustomLayout(t *testing.T) {
	p := NewPluginTimestamp()
	got := runTimestamp(t, p.Process, p.RegisterFlags, []byte("0"), "-layout", "2006-01-02 15:04:05")
	if got != "1970-01-01 00:00:00" {
		t.Fatalf("got %q", got)
	}
}

func TestTimestampRejectsInvalidInput(t *testing.T) {
	p := NewPluginTimestamp()
	_, err := tryTimestamp(p.Process, p.RegisterFlags, []byte("not-a-time"))
	if err == nil || !strings.Contains(err.Error(), "invalid Unix timestamp") {
		t.Fatalf("expected invalid timestamp error, got %v", err)
	}
}

func tryTimestamp(fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) (string, error) {
	fs := flag.NewFlagSet("timestamp", flag.ContinueOnError)
	if registerFlags != nil {
		registerFlags(fs)
	}
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	var out bytes.Buffer
	err := fn(bytes.NewReader(input), &out, fs)
	return out.String(), err
}

func runTimestamp(t *testing.T, fn func(io.Reader, io.Writer, *flag.FlagSet) error, registerFlags func(*flag.FlagSet), input []byte, args ...string) string {
	t.Helper()
	got, err := tryTimestamp(fn, registerFlags, input, args...)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
