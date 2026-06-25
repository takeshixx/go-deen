package compressions

import (
	"flag"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBrotli creates a new brotli plugin (RFC 7932).
func NewPluginBrotli() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "brotli"
	p.Aliases = []string{".brotli", "br", ".br"}
	p.Category = "compressions"
	p.Description = "Brotli is a generic-purpose lossless compression algorithm (RFC 7932)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("level", brotli.DefaultCompression, "compression level (0-11)")
		flags.Int("lgwin", 0, "sliding window size (0-24)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		level := helpers.IntFlag(flags, "level", brotli.DefaultCompression)
		if level < 0 || level > 11 {
			return fmt.Errorf("invalid level %d (must be 0-11)", level)
		}
		lgwin := helpers.IntFlag(flags, "lgwin", 0)
		if lgwin < 0 || lgwin > 24 {
			return fmt.Errorf("invalid window size %d (must be 0-24)", lgwin)
		}
		return compressStream(r, w, func(w io.Writer) (io.WriteCloser, error) {
			return brotli.NewWriterOptions(w, brotli.WriterOptions{Quality: level, LGWin: lgwin}), nil
		})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decompressStream(r, w, func(r io.Reader) (io.Reader, error) {
			return brotli.NewReader(r), nil
		})
	}
	return p
}
