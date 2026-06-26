package codecs

import (
	"bytes"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

// allBytes is every possible byte value, the worst case for byte-safety.
func allBytes() []byte {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

// TestCodecsBinarySafe verifies that reversible codecs round-trip arbitrary
// binary data unchanged. Text-oriented transforms (unicode) are intentionally
// excluded; quoted-printable is covered because its encoder is binary-safe.
func TestCodecsBinarySafe(t *testing.T) {
	plugins := map[string]*types.DeenPlugin{
		"base32":           NewPluginBase32(),
		"base64":           NewPluginBase64(),
		"base85":           NewPluginBase85(),
		"hex":              NewPluginHex(),
		"url":              NewPluginURL(),
		"strconv":          NewPluginStrconv(),
		"rot13":            NewPluginROT13(),
		"quoted-printable": NewPluginQuotedPrintable(),
	}
	input := allBytes()
	for name, p := range plugins {
		encoded, err := tryCodec(p.Process, p.RegisterFlags, input)
		if err != nil {
			t.Errorf("%s: encode failed: %s", name, err)
			continue
		}
		decoded, err := tryCodec(p.Unprocess, p.RegisterFlags, encoded)
		if err != nil {
			t.Errorf("%s: decode failed: %s", name, err)
			continue
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("%s: binary round-trip mismatch (got %d bytes, want %d)", name, len(decoded), len(input))
		}
	}
}
