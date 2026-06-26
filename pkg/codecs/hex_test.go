package codecs

import (
	"encoding/hex"
	"testing"
)

var hexInputData = []byte("asd1239999")
var hexInputDataProcessed = []byte("61736431323339393939")
var hexBinData = "ad13285a5a48976ef51f18c601954a703e3c0c5a"

func TestPluginHexProcess(t *testing.T) {
	p := NewPluginHex()
	assertCodec(t, p, p.Process, hexInputData, hexInputDataProcessed)
}

func TestPluginHexUnprocess(t *testing.T) {
	p := NewPluginHex()
	assertCodec(t, p, p.Unprocess, hexInputDataProcessed, hexInputData)

	decoded, err := hex.DecodeString(hexBinData)
	if err != nil {
		t.Fatal(err)
	}
	assertCodec(t, p, p.Unprocess, []byte(hexBinData), decoded)
}
