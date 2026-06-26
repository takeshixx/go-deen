package codecs

import (
	"bytes"
	"testing"
)

var b32InputData = []byte("deentestdatastringextendedversion321xxx")
var b32InputDataProcessed = []byte("MRSWK3TUMVZXIZDBORQXG5DSNFXGOZLYORSW4ZDFMR3GK4TTNFXW4MZSGF4HQ6A=")
var b32InputDataProcessedHex = []byte("CHIMARJKCLPN8P31EHGN6T3ID5N6EPBOEHIMSP35CHR6ASJJD5NMSCPI65S7GU0=")

func TestPluginBase32Process(t *testing.T) {
	p := NewPluginBase32()
	assertCodec(t, p, p.Process, b32InputData, b32InputDataProcessed)
	assertCodec(t, p, p.Process, b32InputData, b32InputDataProcessedHex, "-hex")
	assertCodec(t, p, p.Process, b32InputData, bytes.TrimSuffix(b32InputDataProcessed, []byte("=")), "-no-pad")
	assertCodec(t, p, p.Process, b32InputData, bytes.TrimSuffix(b32InputDataProcessedHex, []byte("=")), "-hex", "-no-pad")
}

func TestPluginBase32Unprocess(t *testing.T) {
	p := NewPluginBase32()
	assertCodec(t, p, p.Unprocess, b32InputDataProcessed, b32InputData)
	assertCodec(t, p, p.Unprocess, b32InputDataProcessedHex, b32InputData, "-hex")
}

func TestPluginBase32UnprocessInvalid(t *testing.T) {
	p := NewPluginBase32()
	if _, err := tryCodec(p.Unprocess, p.RegisterFlags, b32InputData); err == nil {
		t.Error("decoding non-base32 input should fail")
	}
}
