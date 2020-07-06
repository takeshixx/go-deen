package compressions

import (
	"bytes"
	"io"

	"github.com/ulikunitz/xz/lzma"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginLZMA creates a new PluginLZMA object
func NewPluginLZMA() (p types.DeenPlugin) {
	p.Name = "lzma"
	p.Aliases = []string{".lzma"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressor, err := lzma.NewWriter(&outBuf)
		if err != nil {
			return outBuf.Bytes(), err
		}
		if _, err := io.Copy(compressor, reader); err != nil {
			return outBuf.Bytes(), err
		}
		compressor.Close()
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		decompressor, err := lzma.NewReader(reader)
		if err != nil {
			return outBuf.Bytes(), err
		}
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decompressor); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}

// NewPluginLZMA2 creates a new PluginLZMA2 object
func NewPluginLZMA2() (p types.DeenPlugin) {
	p.Name = "lzma2"
	p.Aliases = []string{".lzma2"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressor, err := lzma.NewWriter2(&outBuf)
		if err != nil {
			return outBuf.Bytes(), err
		}
		if _, err := io.Copy(compressor, reader); err != nil {
			return outBuf.Bytes(), err
		}
		// TODO: maybe always close the writer?
		compressor.Close()
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		decompressor, err := lzma.NewReader2(reader)
		if err != nil {
			return outBuf.Bytes(), err
		}
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decompressor); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	return
}
