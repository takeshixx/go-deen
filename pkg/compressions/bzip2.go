package compressions

import (
	"bytes"
	"compress/bzip2"
	"errors"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBzip2 creates a new zlib plugin
func NewPluginBzip2() (p types.DeenPlugin) {
	p.Name = "bzip2"
	p.Aliases = []string{".bzip2"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return []byte{}, errors.New("Bzip2Process not implemented")
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		decompressor := bzip2.NewReader(reader)
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decompressor); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}
