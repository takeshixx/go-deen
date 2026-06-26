package hashs

import (
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// NewPluginBcrypt creates a plugin
func NewPluginBcrypt() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "bcrypt"
	p.Category = "hashs"
	p.Description = "bcrypt password hashing."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("cost", bcrypt.DefaultCost, "calculation cost")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		cost := helpers.IntFlag(flags, "cost", bcrypt.DefaultCost)
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		out, err := bcrypt.GenerateFromPassword(input, cost)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	return p
}
