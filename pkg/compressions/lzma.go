package compressions

import (
	"flag"
	"io"

	"github.com/ulikunitz/xz/lzma"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginLZMA creates a new LZMA plugin.
func NewPluginLZMA() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "lzma"
	p.Aliases = []string{".lzma"}
	p.Category = "compressions"
	p.Description = "Decoding and encoding of LZMA streams."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return lzma.NewWriter(w)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return lzma.NewReader(r)
		})
	}
	return p
}

// NewPluginLZMA2 creates a new LZMA2 plugin.
func NewPluginLZMA2() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "lzma2"
	p.Aliases = []string{".lzma2"}
	p.Category = "compressions"
	p.Description = "Decoding and encoding of LZMA2 streams."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return lzma.NewWriter2(w)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return lzma.NewReader2(r)
		})
	}
	return p
}
