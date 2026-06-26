package compressions

import (
	"compress/gzip"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginGzip creates a new gzip plugin (RFC 1952).
func NewPluginGzip() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "gzip"
	p.Aliases = []string{".gzip"}
	p.Category = "compressions"
	p.Description = "Implements reading and writing of gzip format compressed files (RFC1952)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("level", gzip.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		level := helpers.IntFlag(flags, "level", gzip.DefaultCompression)
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return gzip.NewWriterLevel(w, level)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		})
	}
	return p
}
