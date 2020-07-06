package codecs

import (
	"bytes"
	"encoding/base32"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBase32 creates a new PluginBase32 object
// Standard base32 encoding, as defined in RFC 4648
func NewPluginBase32() (p types.DeenPlugin) {
	p.Name = "base32"
	p.Aliases = []string{".base32", "b32", ".b32"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := base32.NewEncoder(base32.StdEncoding, &outBuf)
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
		decoder := base32.NewDecoder(base32.StdEncoding, reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}

// NewPluginBase32Hex creates a new PluginBase32Hex object
// “Extended Hex Alphabet” defined in RFC 4648
func NewPluginBase32Hex() (p types.DeenPlugin) {
	p.Name = "base32hex"
	p.Aliases = []string{".base32hex", "b32h", ".b32h"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := base32.NewEncoder(base32.HexEncoding, &outBuf)
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
		decoder := base32.NewDecoder(base32.HexEncoding, reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}
