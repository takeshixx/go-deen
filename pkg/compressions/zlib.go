package compressions

import (
	"compress/zlib"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginZlib creates a new zlib plugin (RFC 1950).
func NewPluginZlib() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "zlib"
	p.Aliases = []string{".zlib"}
	p.Category = "compressions"
	p.Description = "Implements reading and writing of zlib format compressed data (RFC1950)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("level", zlib.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		level := helpers.IntFlag(flags, "level", zlib.DefaultCompression)
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return zlib.NewWriterLevel(w, level)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return zlib.NewReader(r)
		})
	}
	return p
}
