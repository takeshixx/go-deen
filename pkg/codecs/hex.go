package codecs

import (
	"encoding/hex"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHex creates a new ASCII hex plugin.
func NewPluginHex() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "hex"
	p.Aliases = []string{".hex", "asciihex", ".asciihex"}
	p.Category = "codecs"
	p.Description = "Apply ASCII hex encoding or decoding to data."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		_, err := io.Copy(hex.NewEncoder(w), r)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return decodeTrimmed(r, w, func(r io.Reader) io.Reader { return hex.NewDecoder(r) })
	}
	return p
}
