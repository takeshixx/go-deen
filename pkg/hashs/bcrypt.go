package hashs

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// NewPluginBcrypt creates a plugin
func NewPluginBcrypt() (p types.DeenPlugin) {
	p.Name = "bcrypt"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		var inBuf bytes.Buffer
		_, err = io.Copy(&inBuf, reader)
		if err != nil {
			return *new([]byte), err
		}
		outBuf, err := bcrypt.GenerateFromPassword(inBuf.Bytes(), bcrypt.DefaultCost)
		return outBuf, err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		costFlag := flags.Lookup("cost")
		cost, err := strconv.Atoi(costFlag.Value.String())
		if err != nil {
			cost = bcrypt.DefaultCost
		}
		var inBuf bytes.Buffer
		_, err = io.Copy(&inBuf, reader)
		if err != nil {
			return *new([]byte), err
		}
		outBuf, err := bcrypt.GenerateFromPassword(inBuf.Bytes(), cost)
		return outBuf, err
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		bcryptCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		bcryptCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "bcrypt password hashing.\n\n")
			bcryptCmd.PrintDefaults()
		}
		bcryptCmd.Int("cost", 10, "calculation cost")
		bcryptCmd.Parse(args)
		return bcryptCmd
	}
	return
}
