package codecs

import (
	"bytes"
	"testing"
)

var b64InputData = []byte("asd123<<<<>>>>deentestdata23xxxx")
var b64InputDataProcessed = []byte("YXNkMTIzPDw8PD4+Pj5kZWVudGVzdGRhdGEyM3h4eHg=")
var b64InputDataProcessedURL = []byte("YXNkMTIzPDw8PD4-Pj5kZWVudGVzdGRhdGEyM3h4eHg=")

func TestPluginBase64Process(t *testing.T) {
	p := NewPluginBase64()
	assertCodec(t, p, p.Process, b64InputData, b64InputDataProcessed)
	assertCodec(t, p, p.Process, b64InputData, b64InputDataProcessedURL, "-url")
	assertCodec(t, p, p.Process, b64InputData, bytes.ReplaceAll(b64InputDataProcessed, []byte("="), nil), "-raw")
	assertCodec(t, p, p.Process, b64InputData, bytes.ReplaceAll(b64InputDataProcessedURL, []byte("="), nil), "-raw", "-url")
}

func TestPluginBase64Unprocess(t *testing.T) {
	p := NewPluginBase64()
	// Default decode auto-detects the alphabet.
	assertCodec(t, p, p.Unprocess, b64InputDataProcessed, b64InputData)
	assertCodec(t, p, p.Unprocess, b64InputDataProcessedURL, b64InputData)
	assertCodec(t, p, p.Unprocess, bytes.ReplaceAll(b64InputDataProcessed, []byte("="), nil), b64InputData)
	// Explicit flags.
	assertCodec(t, p, p.Unprocess, b64InputDataProcessedURL, b64InputData, "-url")
	assertCodec(t, p, p.Unprocess, b64InputDataProcessed, b64InputData, "-strict")
}
