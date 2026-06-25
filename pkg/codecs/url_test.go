package codecs

import "testing"

var urlInputData = []byte("test?deen=true")
var urlInputDataProcessed = []byte("test%3Fdeen%3Dtrue")

func TestPluginURLProcess(t *testing.T) {
	p := NewPluginURL()
	assertCodec(t, p, p.Process, urlInputData, urlInputDataProcessed)
}

func TestPluginURLUnprocess(t *testing.T) {
	p := NewPluginURL()
	assertCodec(t, p, p.Unprocess, urlInputDataProcessed, urlInputData)
}
