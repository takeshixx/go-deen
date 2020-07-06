package compressions

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginZlib creates a new zlib plugin
func NewPluginZlib() (p types.DeenPlugin) {
	p.Name = "zlib"
	p.Aliases = []string{".zlib"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressor := zlib.NewWriter(&outBuf)
		if _, err := io.Copy(compressor, reader); err != nil {
			return outBuf.Bytes(), err
		}
		compressor.Close()
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		decompressor, err := zlib.NewReader(reader)
		if err != nil {
			return outBuf.Bytes(), err
		}
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decompressor); err != nil {
			return outBuf.Bytes(), err
		}
		return outBuf.Bytes(), err
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		zlibCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		zlibCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Implements reading and writing of zlib format compressed data (RFC1950).\n\n")
			zlibCmd.PrintDefaults()
		}
		zlibCmd.Parse(args)
		return zlibCmd
	}
	return
}
