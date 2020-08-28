package hashs

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSHA224 creates a plugin
func NewPluginSHA224() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "sha224"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha256.New224()
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
			fmt.Fprintf(os.Stderr, "SHA2 is a set of cryptographic hash functions designed\nby the United States National Security Agency (NSA).\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginSHA256 creates a plugin
func NewPluginSHA256() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "sha256"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha256.New()
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
			fmt.Fprintf(os.Stderr, "SHA2 is a set of cryptographic hash functions designed\nby the United States National Security Agency (NSA).\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginSHA384 creates a plugin
func NewPluginSHA384() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "sha384"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha512.New384()
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
			fmt.Fprintf(os.Stderr, "SHA2 is a set of cryptographic hash functions designed\nby the United States National Security Agency (NSA).\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginSHA512 creates a plugin
func NewPluginSHA512() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "sha512"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha512.New()
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
			fmt.Fprintf(os.Stderr, "SHA2 is a set of cryptographic hash functions designed\nby the United States National Security Agency (NSA).\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
