package hashs

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/scrypt"
)

// NewPluginScrypt creates a new plugin
func NewPluginScrypt() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "scrypt"
	p.Category = "hashs"
	p.Description = "scrypt key derivation function as defined in Colin Percival's paper\n\"Stronger Key Derivation via Sequential Memory-Hard Functions\"."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("salt", "", "hex encoded string used as salt")
		flags.Int("len", 32, "output key length")
		flags.Int("cost", 1<<15, "calculation cost")
		flags.Int("r", 8, "parallelization parameter")
		flags.Int("p", 1, "blocksize parameter")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		cost := helpers.IntFlag(flags, "cost", 1<<15)
		length := helpers.IntFlag(flags, "len", 32)
		rParam := helpers.IntFlag(flags, "r", 8)
		pParam := helpers.IntFlag(flags, "p", 1)
		var salt []byte
		if s := helpers.StringFlag(flags, "salt"); s != "" {
			decoded, err := hex.DecodeString(s)
			if err != nil {
				return err
			}
			salt = decoded
		}
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		key, err := scrypt.Key(input, salt, cost, rParam, pParam, length)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, base64.StdEncoding.EncodeToString(key))
		return err
	}
	return p
}
