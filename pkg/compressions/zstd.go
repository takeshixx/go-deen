package compressions

import (
	"flag"
	"io"

	"github.com/klauspost/compress/zstd"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginZstd creates a new Zstandard plugin (RFC 8878).
func NewPluginZstd() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "zstd"
	p.Aliases = []string{".zstd", "zst", ".zst"}
	p.Category = "compressions"
	p.Description = "Zstandard compression (RFC 8878)."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return zstd.NewWriter(w)
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return zstd.NewReader(r)
		})
	}
	return p
}
