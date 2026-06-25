package compressions

import (
	"flag"
	"fmt"
	"io"

	"github.com/dsnet/compress/bzip2"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBzip2 creates a new bzip2 plugin.
func NewPluginBzip2() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "bzip2"
	p.Aliases = []string{".bzip2"}
	p.Category = "compressions"
	p.Description = "BZip2 compressed data format."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("level", bzip2.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		level := helpers.IntFlag(flags, "level", bzip2.DefaultCompression)
		if level < bzip2.BestSpeed || level > bzip2.BestCompression {
			return fmt.Errorf("invalid level %d (must be %d-%d)", level, bzip2.BestSpeed, bzip2.BestCompression)
		}
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return bzip2.NewWriter(w, &bzip2.WriterConfig{Level: level})
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return bzip2.NewReader(r, nil)
		})
	}
	return p
}
