package codecs

import "testing"

var strconvTestData = []byte("☺")
var strconvTestDataProcessed = []byte("\\u263a")

func TestPluginStrconvProcess(t *testing.T) {
	p := NewPluginStrconv()
	assertCodec(t, p, p.Process, strconvTestData, strconvTestDataProcessed)
}

func TestPluginStrconvUnprocess(t *testing.T) {
	p := NewPluginStrconv()
	assertCodec(t, p, p.Unprocess, strconvTestDataProcessed, strconvTestData)
}
