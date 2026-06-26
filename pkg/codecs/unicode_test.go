package codecs

import (
	"bytes"
	"testing"
)

func TestPluginUnicodeRoundTrip(t *testing.T) {
	input := []byte("deen unicode test 123")
	for _, command := range []string{"utf8", "utf16", "utf32", "euckr"} {
		p := NewPluginUnicode()
		p.Command = command
		encoded := runCodec(t, p.Process, p.RegisterFlags, input)
		decoded := runCodec(t, p.Unprocess, p.RegisterFlags, encoded)
		if !bytes.Equal(decoded, input) {
			t.Errorf("%s round-trip mismatch: got %q, want %q", command, decoded, input)
		}
	}
}
