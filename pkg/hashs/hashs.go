package hashs

import (
	"encoding/hex"
	"flag"
	"hash"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// hashPlugin builds a one-way hash plugin that streams the input through the
// hash produced by newHash and writes the lowercase hex digest to the output.
func hashPlugin(name, description string, aliases []string, newHash func() hash.Hash) *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = name
	p.Aliases = aliases
	p.Category = "hashs"
	p.Description = description
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		h := newHash()
		if _, err := io.Copy(h, r); err != nil {
			return err
		}
		_, err := io.WriteString(w, hex.EncodeToString(h.Sum(nil)))
		return err
	}
	return p
}
