package compressions

import (
	"compress/flate"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginFlate creates a new DEFLATE plugin (RFC 1951).
func NewPluginFlate() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "flate"
	p.Aliases = []string{".flate"}
	p.Category = "compressions"
	p.Description = "Implements the DEFLATE compressed data format (RFC1951)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("level", flate.DefaultCompression, "compression level (-1 default, 0 none, 1 best speed, 9 best compression)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		level := helpers.IntFlag(flags, "level", flate.DefaultCompression)
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, level)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return flate.NewReader(r), nil
		})
	}
	return p
}
