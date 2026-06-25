package codecs

import "testing"

var b85InputData = []byte("asd1239999")
var b85InputDataProcessed = []byte("@<5s61,CpN3B7")

func TestPluginBase85Process(t *testing.T) {
	p := NewPluginBase85()
	assertCodec(t, p, p.Process, b85InputData, b85InputDataProcessed)
}

func TestPluginBase85Unprocess(t *testing.T) {
	p := NewPluginBase85()
	assertCodec(t, p, p.Unprocess, b85InputDataProcessed, b85InputData)
}
