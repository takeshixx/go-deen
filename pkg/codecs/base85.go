package codecs

import (
	"bytes"
	"encoding/ascii85"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBase85 creates a new PluginBase85 object
func NewPluginBase85() (p types.DeenPlugin) {
	p.Name = "base85"
	p.Aliases = []string{".base85", "b85", ".b85",
		"ascii85", ".ascii85", "a85",
		".a85"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := ascii85.NewEncoder(&outBuf)
		if _, err := io.Copy(encoder, reader); err != nil {
			return outBuf.Bytes(), err
		}
		encoder.Close()
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		// We have to remove leading/trailing whitespaces
		wrappedReader := trimReader{}
		wrappedReader.rd = reader
		decoder := ascii85.NewDecoder(reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}
