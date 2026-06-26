package codecs

import (
	"encoding/ascii85"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBase85 creates a new ascii85 plugin.
func NewPluginBase85() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "base85"
	p.Aliases = []string{".base85", "b85", ".b85", "ascii85", ".ascii85", "a85", ".a85"}
	p.Category = "codecs"
	p.Description = "Implements the ascii85 data encoding as used in the btoa tool and\nAdobe's PostScript and PDF document formats."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return encodeStream(r, w, func(w io.Writer) io.WriteCloser { return ascii85.NewEncoder(w) })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decodeTrimmed(r, w, func(r io.Reader) io.Reader { return ascii85.NewDecoder(r) })
	}
	return p
}
