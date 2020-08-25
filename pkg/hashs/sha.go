package hashs

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSHA1 creates a plugin
func NewPluginSHA1() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "sha1"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha1.New()
		if _, err := io.Copy(hasher, reader); err != nil {
			return *new([]byte), err
		}
		hashSum := hasher.Sum(nil)
		outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
		_ = hex.Encode(outBuf, hashSum[:])
		return outBuf, err
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "SHA1 is a cryptographic hash function which takes an input\nand produces a 160-bit (20-byte) hash value known as a\nmessage digest.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
