package compressions

import (
	"compress/lzw"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginLzw creates a new LZW plugin.
func NewPluginLzw() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "lzw"
	p.Aliases = []string{".lzw"}
	p.Category = "compressions"
	p.Description = "Implements the Lempel-Ziv-Welch compressed data format."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("order", int(lzw.LSB), "0 = LSB (GIF), 1 = MSB (TIFF & PDF)")
		flags.Int("lit-width", 8, "number of bits per code, 2-8")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		order := lzw.Order(helpers.IntFlag(flags, "order", int(lzw.LSB)))
		width := helpers.IntFlag(flags, "lit-width", 8)
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return lzw.NewWriter(w, order, width), nil
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		order := lzw.Order(helpers.IntFlag(flags, "order", int(lzw.LSB)))
		width := helpers.IntFlag(flags, "lit-width", 8)
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return lzw.NewReader(r, order, width), nil
		})
	}
	return p
}
