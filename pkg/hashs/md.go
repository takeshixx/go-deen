package hashs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginMD4 creates a plugin
func NewPluginMD4() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "md4"
	p.Aliases = []string{}
	p.Category = "hashs"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			hasher := md4.New()
			_, err := io.Copy(hasher, task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into encoder in MD4 failed")
			}
			hashSum := hasher.Sum(nil)
			encodedBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
			_ = hex.Encode(encodedBuf, hashSum[:])
			outBuf := bytes.NewBuffer(encodedBuf)
			_, err = io.Copy(task.PipeWriter, outBuf)
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in MD4 failed")
			}
		}()
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "MD4 Message-Digest Algorithm is a cryptographic hash\nfunction with a digest length of 128 bits.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginMD5 creates a plugin
func NewPluginMD5() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "md5"
	p.Aliases = []string{}
	p.Category = "hashs"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := md5.New()
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
			fmt.Fprintf(os.Stderr, "MD5 Message-Digest Algorithm is a cryptographic hash\nfunction with a digest length of 128 bits.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginRIPEMD160 creates a plugin
func NewPluginRIPEMD160() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "ripemd160"
	p.Aliases = []string{"md160"}
	p.Category = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := ripemd160.New()
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
			fmt.Fprintf(os.Stderr, "RIPEMD (RIPE Message Digest) is a family of cryptographic\nhash functions developed in 1992.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
