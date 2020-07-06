package compressions

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginGzip creates a new zlib plugin
func NewPluginGzip() (p types.DeenPlugin) {
	p.Name = "gzip"
	p.Aliases = []string{".gzip"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressor := gzip.NewWriter(&outBuf)
		if _, err := io.Copy(compressor, reader); err != nil {
			return outBuf.Bytes(), err
		}
		compressor.Close()
		return outBuf.Bytes(), err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		decompressor, err := gzip.NewReader(reader)
		if err != nil {
			return outBuf.Bytes(), err
		}
		wrapper := struct{ io.Writer }{&outBuf}
		if _, err := io.Copy(wrapper, decompressor); err != nil {
			return outBuf.Bytes(), err
		}
		decompressor.Close()
		return outBuf.Bytes(), err
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		gzipCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		gzipCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Implements reading and writing of gzip format compressed files (RFC1952).\n\n")
			gzipCmd.PrintDefaults()
		}
		gzipCmd.Parse(args)
		return gzipCmd
	}
	return
}
