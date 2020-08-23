package hashs

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"lukechampine.com/blake3"

	"github.com/takeshixx/deen/pkg/types"
)

func doBLAKE2x(reader *io.Reader, macKey []byte, length uint16) ([]byte, error) {
	var hasher blake2s.XOF
	var err error
	if len(macKey) != 0 {
		hasher, err = blake2s.NewXOF(length, macKey)
	} else {
		hasher, err = blake2s.NewXOF(length, nil)
	}
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(hasher, *reader); err != nil {
		return nil, err
	}
	var outBuf bytes.Buffer
	hexEncoder := hex.NewEncoder(&outBuf)
	_, err = io.Copy(hexEncoder, hasher)
	if err != nil {
		return nil, err
	}
	return outBuf.Bytes(), err
}

// NewPluginBLAKE2x creates a plugin
func NewPluginBLAKE2x() (p types.DeenPlugin) {
	p.Name = "blake2x"
	p.Aliases = []string{"b2x"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2x(&reader, nil, uint16(32))
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		var key []byte
		if macKey.Value.String() != "" {
			key = []byte(macKey.Value.String())
		} else {
			key = nil
		}
		lengthFlag := flags.Lookup("len")
		length, err := strconv.Atoi(lengthFlag.Value.String())
		if err != nil {
			return nil, err
		}
		if length < 1 || length > 65535 {
			length = 32
		}
		return doBLAKE2x(&reader, key, uint16(length))
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "BLAKE2 is a fast and secure cryptographic hash function defined \nin RFC 7693.\n\n")
			fmt.Fprintf(os.Stderr, "BLAKE2X is a family of extensible-output functions (XOFs). Whereas BLAKE2 is limited to 64-byte digests, BLAKE2X allows for digests of up to 256 GiB.\n\n")
			flags.PrintDefaults()
		}
		flags.String("key", "", "MAC key")
		flags.Int("len", 32, "length of the output hash in bytes, must be between 1 and 65535")
		flags.Parse(args)
		return flags
	}
	return
}

func doBLAKE2s(reader *io.Reader, macKey []byte, length int) ([]byte, error) {
	var err error
	var hasher hash.Hash
	if length == 32 {
		if len(macKey) != 0 {
			hasher, err = blake2s.New256(macKey)
		} else {
			hasher, err = blake2s.New256(nil)
		}
	} else {
		if len(macKey) != 0 {
			hasher, err = blake2s.New128(macKey)
		} else {
			return []byte(""), errors.New("BLAKE2s128 requres a key")
		}
	}
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(hasher, *reader); err != nil {
		return nil, err
	}
	hashSum := hasher.Sum(nil)
	outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
	_ = hex.Encode(outBuf, hashSum[:])
	return outBuf, err
}

// NewPluginBLAKE2s creates a plugin
func NewPluginBLAKE2s() (p types.DeenPlugin) {
	p.Name = "blake2s"
	p.Aliases = []string{"b2s"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2s(&reader, nil, 32)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		var key []byte
		if macKey.Value.String() != "" {
			key = []byte(macKey.Value.String())
		} else {
			key = nil
		}
		lengthFlag := flags.Lookup("len")
		length, err := strconv.Atoi(lengthFlag.Value.String())
		if err != nil {
			return nil, err
		}
		if length != 16 && length != 32 {
			return nil, errors.New("Invalid length")
		}
		return doBLAKE2s(&reader, key, length)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "BLAKE2 is a fast and secure cryptographic hash function defined \nin RFC 7693.\n\n")
			fmt.Fprintf(os.Stderr, "BLAKE2s is optimized for 8- to 32-bit platforms and produces\ndigests of any size between 1 and 32 bytes.\n\n")
			flags.PrintDefaults()
		}
		flags.String("key", "", "MAC key")
		flags.Int("len", 32, "length of the output hash in bytes, must be either 16 or 32")
		flags.Parse(args)
		return flags
	}
	return
}

func doBLAKE2b(reader *io.Reader, macKey []byte, length int) ([]byte, error) {
	var hasher hash.Hash
	var err error
	if len(macKey) != 0 {
		hasher, err = blake2b.New(length, macKey)
	} else {
		hasher, err = blake2b.New(length, nil)
	}
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(hasher, *reader); err != nil {
		return nil, err
	}
	hashSum := hasher.Sum(nil)
	outBuf := []byte(fmt.Sprintf("%x", hashSum))
	return outBuf, err
}

// NewPluginBLAKE2b creates a plugin
func NewPluginBLAKE2b() (p types.DeenPlugin) {
	p.Name = "blake2b"
	p.Aliases = []string{"b2b"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE2b(&reader, nil, 64)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		macKey := flags.Lookup("key")
		var key []byte
		if macKey.Value.String() != "" {
			key = []byte(macKey.Value.String())
		} else {
			key = nil
		}
		lengthFlag := flags.Lookup("len")
		length, err := strconv.Atoi(lengthFlag.Value.String())
		if err != nil {
			return nil, err
		}
		if length < 1 || length > 64 {
			length = 32
		}
		return doBLAKE2b(&reader, key, length)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "BLAKE2 is a fast and secure cryptographic hash function defined \nin RFC 7693.\n\n")
			fmt.Fprintf(os.Stderr, "BLAKE2b (or just BLAKE2) is optimized for 64-bit platforms and\nproduces digests of any size between 1 and 64 bytes . This\nplugin generates 512 bit hashs.\n\n")
			flags.PrintDefaults()
		}
		flags.String("key", "", "MAC key")
		flags.Int("len", 32, "length of the output hash in bytes, must be between 1 and 64")
		flags.Parse(args)
		return flags
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
		return nil, err
	}
	hashSum := hasher.Sum(nil)
	outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
	_ = hex.Encode(outBuf, hashSum[:])
	return outBuf, err
}

// NewPluginBLAKE3 creates a plugin
func NewPluginBLAKE3() (p types.DeenPlugin) {
	p.Name = "blake3"
	p.Aliases = []string{"b3"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doBLAKE3(32, &reader, []byte{}, false, "")
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		length := flags.Lookup("length")
		outLen, err := strconv.Atoi(length.Value.String())
		if err != nil {
			return nil, err
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
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s: \n\n", p.Name)
			fmt.Fprintf(os.Stderr, "BLAKE3 is a cryptographic hash function that is fast and secure.\n")
			fmt.Fprintf(os.Stderr, "It can be used as PRF, MAC, KDF, and XOF as well as a regular hash.\n\n")
			flags.PrintDefaults()
		}
		flags.String("key", "", "key (requires 32 bytes)")
		flags.Int("length", 32, "number of output bytes")
		flags.String("derive-key", "", "derive key")
		flags.String("context", "", "context for key derivation")
		flags.Parse(args)
		return flags
	}
	return
}
