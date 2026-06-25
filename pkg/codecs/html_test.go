package codecs

import "testing"

var htmlInputData = []byte("<h1>deen</h1>")
var htmlInputDataProcessed = []byte("&lt;h1&gt;deen&lt;/h1&gt;")

func TestPluginHTMLProcess(t *testing.T) {
	p := NewPluginHTML()
	assertCodec(t, p, p.Process, htmlInputData, htmlInputDataProcessed)
}

func TestPluginHTMLUnprocess(t *testing.T) {
	p := NewPluginHTML()
	assertCodec(t, p, p.Unprocess, htmlInputDataProcessed, htmlInputData)
}
