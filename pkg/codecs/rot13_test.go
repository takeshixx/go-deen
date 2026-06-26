package codecs

import (
	"bytes"
	"testing"
)

func TestPluginROT13(t *testing.T) {
	p := NewPluginROT13()
	input := []byte("Hello, deen 123!")
	encoded := []byte("Uryyb, qrra 123!")
	assertCodec(t, p, p.Process, input, encoded)
	// ROT13 is its own inverse.
	assertCodec(t, p, p.Unprocess, encoded, input)
	// Double application returns the original.
	once := runCodec(t, p.Process, p.RegisterFlags, input)
	twice := runCodec(t, p.Process, p.RegisterFlags, once)
	if !bytes.Equal(twice, input) {
		t.Errorf("ROT13 applied twice should be identity, got %q", twice)
	}
}

func TestPluginQuotedPrintable(t *testing.T) {
	p := NewPluginQuotedPrintable()
	input := []byte("equals = sign and ümlaut")
	encoded := runCodec(t, p.Process, p.RegisterFlags, input)
	if bytes.Equal(encoded, input) {
		t.Error("quoted-printable encoding produced identical output")
	}
	decoded := runCodec(t, p.Unprocess, p.RegisterFlags, encoded)
	if !bytes.Equal(decoded, input) {
		t.Errorf("quoted-printable round-trip mismatch: got %q, want %q", decoded, input)
	}
}
