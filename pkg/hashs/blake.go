package hashs

import (
	"encoding/hex"
	"errors"
	"flag"
	"hash"
	"io"
	"strconv"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"lukechampine.com/blake3"

	"github.com/takeshixx/deen/pkg/types"
)

// Generic function that calculates BLAKE2 variants
func doBLAKE2(variant string, reader *io.Reader, macKey []byte) ([]byte, error) {
	var hasher hash.Hash
	var err error
	switch variant {
	case "blake2s-256":
		if len(macKey) != 0 {
			hasher, err = blake2s.New256(macKey)
		} else {
			hasher, err = blake2s.New256(nil)
		}
	case "blake2s-128":
		if len(macKey) != 0 {
			hasher, err = blake2s.New128(macKey)
		} else {
			hasher, err = blake2s.New128(nil)
		}
	case "blake2b-512":
		if len(macKey) != 0 {
			hasher, err = blake2b.New512(macKey)
		} else {
			hasher, err = blake2b.New512(nil)
		}
	case "blake2b-384":
		if len(macKey) != 0 {
			hasher, err = blake2b.New384(macKey)
		} else {
			hasher, err = blake2b.New384(nil)
		}
	case "blake2b-256":
		if len(macKey) != 0 {
			hasher, err = blake2b.New256(macKey)
		} else {
			hasher, err = blake2b.New256(nil)
		}
	default:
		return *new([]byte), err
	}
	if err != nil {
		return *new([]byte), err
	}
	if _, err := io.Copy(hasher, *reader); err != nil {
		return *new([]byte), err
	}
	hashSum := hasher.Sum(nil)
	outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
	_ = hex.Encode(outBuf, hashSum[:])
	return outBuf, err
}

// NewPluginBLAKE2s creates a plugin
func NewPluginBLAKE2s() (p types.DeenPlugin) {
	p.Name = "blake2s"
	p.Aliases = []string{"blake2s256", "blake2s-256"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2("blake2s-256", &reader, []byte{})
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		if macKey.Value.String() != "" {
			return doBLAKE2("blake2s-256", &reader, []byte(macKey.Value.String()))
		}
		return doBLAKE2("blake2s-256", &reader, []byte{})
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "MAC key")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}

// NewPluginBLAKE2s128 creates a plugin
func NewPluginBLAKE2s128() (p types.DeenPlugin) {
	p.Name = "blake2s128"
	p.Aliases = []string{"blake2s-128"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2("blake2s-128", &reader, []byte{})
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		if macKey.Value.String() != "" {
			return doBLAKE2("blake2s-128", &reader, []byte(macKey.Value.String()))
		}
		return doBLAKE2("blake2s-128", &reader, []byte{})
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "MAC key")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}

// NewPluginBLAKE2b creates a plugin
func NewPluginBLAKE2b() (p types.DeenPlugin) {
	p.Name = "blake2b"
	p.Aliases = []string{"blake2b512", "blake2b-512"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2("blake2b-512", &reader, []byte{})
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		if macKey.Value.String() != "" {
			return doBLAKE2("blake2b-512", &reader, []byte(macKey.Value.String()))
		}
		return doBLAKE2("blake2b-512", &reader, []byte{})
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "MAC key")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}

// NewPluginBLAKE2b384 creates a plugin
func NewPluginBLAKE2b384() (p types.DeenPlugin) {
	p.Name = "blake2b384"
	p.Aliases = []string{"blake2b-384"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2("blake2b-384", &reader, []byte{})
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		if macKey.Value.String() != "" {
			return doBLAKE2("blake2b-384", &reader, []byte(macKey.Value.String()))
		}
		return doBLAKE2("blake2b-384", &reader, []byte{})
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "MAC key")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}

// NewPluginBLAKE2b256 creates a plugin
func NewPluginBLAKE2b256() (p types.DeenPlugin) {
	p.Name = "blake2b256"
	p.Aliases = []string{"blake2b-256"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2("blake2b-256", &reader, []byte{})
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		if macKey.Value.String() != "" {
			return doBLAKE2("blake2b-256", &reader, []byte(macKey.Value.String()))
		}
		return doBLAKE2("blake2b-256", &reader, []byte{})
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "MAC key")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}

func doBLAKE3(outLen int, reader *io.Reader, key []byte, derive bool, ctx string) ([]byte, error) {
	var hasher hash.Hash
	var err error
	if len(key) > 0 {
		if derive {
			derivedKey := make([]byte, len(key))
			blake3.DeriveKey(derivedKey, ctx, key)
			buf := make([]byte, hex.EncodedLen(len(derivedKey[:])))
			_ = hex.Encode(buf, derivedKey[:])
			return buf, err
		}
		hasher = blake3.New(outLen, key)
	} else {
		hasher = blake3.New(outLen, nil)
	}
	if _, err := io.Copy(hasher, *reader); err != nil {
		return *new([]byte), err
	}
	hashSum := hasher.Sum(nil)
	outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
	_ = hex.Encode(outBuf, hashSum[:])
	return outBuf, err
}

// NewPluginBLAKE3 creates a plugin
func NewPluginBLAKE3() (p types.DeenPlugin) {
	p.Name = "blake3"
	p.Aliases = []string{"blake3512", "blake3-512"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE3(32, &reader, []byte{}, false, "")
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		length := flags.Lookup("length")
		outLen, err := strconv.Atoi(length.Value.String())
		if err != nil {
			return *new([]byte), err
		}
		if outLen == 0 {
			outLen = 32
		} else if outLen != 32 && outLen != 64 {
			outLen = 32
		}
		rawKey := flags.Lookup("key")
		deriveKey := flags.Lookup("derive-key")
		context := flags.Lookup("context")
		if deriveKey.Value.String() != "" {
			var ctx string
			if context.Value.String() != "" {
				ctx = context.Value.String()
			} else {
				ctx = "BLAKE3 2020-02-13 13:33:37 test data context"
			}
			return doBLAKE3(outLen, &reader, []byte(deriveKey.Value.String()), true, ctx)
		} else if rawKey.Value.String() != "" {
			if len(rawKey.Value.String()) != 32 {
				return *new([]byte), errors.New("Invalid key length")
			}
			return doBLAKE3(outLen, &reader, []byte(rawKey.Value.String()), false, "")
		}
		return doBLAKE3(outLen, &reader, []byte{}, false, "")
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		blakeCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		blakeCmd.String("key", "", "key (requires 32 bytes)")
		blakeCmd.Int("length", 32, "number of output bytes")
		blakeCmd.String("derive-key", "", "derive key")
		blakeCmd.String("context", "", "context for key derivation")
		blakeCmd.Parse(args)
		return blakeCmd
	}
	return
}
