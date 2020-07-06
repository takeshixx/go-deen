package codecs

import (
	"bytes"
	"encoding/base64"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBase64 creates a new PluginBase64 object
func NewPluginBase64() (p types.DeenPlugin) {
	p.Name = "base64"
	p.Aliases = []string{".base64", "b64", ".b64"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := base64.NewEncoder(base64.StdEncoding, &outBuf)
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
		decoder := base64.NewDecoder(base64.StdEncoding, reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}

// NewPluginBase64Url creates a new PluginBase64Url object
func NewPluginBase64Url() (p types.DeenPlugin) {
	p.Name = "base64url"
	p.Aliases = []string{".base64url", "b64u", ".b64u"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := base64.NewEncoder(base64.URLEncoding, &outBuf)
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
		decoder := base64.NewDecoder(base64.URLEncoding, reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}
