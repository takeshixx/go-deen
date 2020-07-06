package codecs

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHex creates a new PluginHex object
func NewPluginHex() (p types.DeenPlugin) {
	p.Name = "hex"
	p.Aliases = []string{".hex", "asciihex", ".asciihex"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		encoder := hex.NewEncoder(&outBuf)
		if _, err := io.Copy(encoder, reader); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		// We have to remove leading/trailing whitespaces
		wrappedReader := trimReader{}
		wrappedReader.rd = reader
		decoder := hex.NewDecoder(wrappedReader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decoder); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}

type trimReader struct {
	rd io.Reader
}

func (tr trimReader) Read(buf []byte) (int, error) {
	n, err := tr.rd.Read(buf)
	t := bytes.TrimSpace(buf[:n])
	n = copy(buf, t)
	return n, err
}
