package hashs

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// NewPluginScrypt creates a new plugin
func NewPluginScrypt() (p types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "scrypt"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var inBuf bytes.Buffer
		_, err := io.Copy(&inBuf, reader)
		if err != nil {
			return nil, err
		}
		tempBuf, err := scrypt.Key(inBuf.Bytes(), nil, 1<<15, 8, 1, 32)
		if err != nil {
			return nil, err
		}
		outBuf := []byte(base64.StdEncoding.EncodeToString(tempBuf))
		return outBuf, err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		costFlag := flags.Lookup("cost")
		cost, err := strconv.Atoi(costFlag.Value.String())
		if err != nil {
			cost = bcrypt.DefaultCost
		}

		lenFlag := flags.Lookup("len")
		length, err := strconv.Atoi(lenFlag.Value.String())
		if err != nil {
			return nil, err
		}

		rFlag := flags.Lookup("r")
		rParam, err := strconv.Atoi(rFlag.Value.String())
		if err != nil {
			return nil, err
		}

		pFlag := flags.Lookup("p")
		pParam, err := strconv.Atoi(pFlag.Value.String())
		if err != nil {
			return nil, err
		}

		saltFlag := flags.Lookup("salt")
		salt, err := hex.DecodeString(saltFlag.Value.String())
		if err != nil {
			return nil, err
		}

		var inBuf bytes.Buffer
		var outBuf []byte
		_, err = io.Copy(&inBuf, reader)
		if err != nil {
			return nil, err
		}

		tempBuf, err := scrypt.Key(inBuf.Bytes(), salt, cost, rParam, pParam, length)
		if err != nil {
			return nil, err
		}
		outBuf = []byte(base64.StdEncoding.EncodeToString(tempBuf))
		return outBuf, err
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "scrypt key derivation function as defined in Colin Percival's\npaper \"Stronger Key Derivation via Sequential Memory-Hard\nFunctions\".\n\n")
			flags.PrintDefaults()
		}
		flags.String("salt", "", "hex encoded string used as salt")
		flags.Int("len", 32, "output key length")
		flags.Int("cost", 1<<15, "calculation cost")
		flags.Int("r", 8, "parallelization parameter")
		flags.Int("p", 1, "blocksize parameter")
		flags.Parse(args)
		return flags
	}
	return
}
