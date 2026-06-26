package hashs

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"lukechampine.com/blake3"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

const blake2Intro = "BLAKE2 is a fast and secure cryptographic hash function defined in\nRFC 7693."

func macKeyFlag(flags *flag.FlagSet) []byte {
	if k := helpers.StringFlag(flags, "key"); k != "" {
		return []byte(k)
	}
	return nil
}

func writeHex(w io.Writer, data []byte) error {
	_, err := io.WriteString(w, hex.EncodeToString(data))
	return err
}

func doBLAKE2x(r io.Reader, w io.Writer, macKey []byte, length uint16) error {
	hasher, err := blake2s.NewXOF(length, macKey)
	if err != nil {
		return err
	}
	if _, err := io.Copy(hasher, r); err != nil {
		return err
	}
	hexEncoder := hex.NewEncoder(w)
	_, err = io.Copy(hexEncoder, hasher)
	return err
}

// NewPluginBLAKE2x creates a plugin
func NewPluginBLAKE2x() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "blake2x"
	p.Aliases = []string{"b2x"}
	p.Category = "hashs"
	p.Description = blake2Intro + "\n\nBLAKE2X is a family of extensible-output functions (XOFs)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("key", "", "MAC key")
		flags.Int("len", 32, "length of the output hash in bytes, must be between 1 and 65535")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		length := helpers.IntFlag(flags, "len", 32)
		if length < 1 || length > 65535 {
			length = 32
		}
		return doBLAKE2x(r, w, macKeyFlag(flags), uint16(length))
	}
	return p
}

func doBLAKE2s(r io.Reader, w io.Writer, macKey []byte, length int) error {
	var hasher hash.Hash
	var err error
	if length == 32 {
		hasher, err = blake2s.New256(macKey)
	} else {
		if len(macKey) == 0 {
			return errors.New("BLAKE2s128 requires a key")
		}
		hasher, err = blake2s.New128(macKey)
	}
	if err != nil {
		return err
	}
	if _, err := io.Copy(hasher, r); err != nil {
		return err
	}
	return writeHex(w, hasher.Sum(nil))
}

// NewPluginBLAKE2s creates a plugin
func NewPluginBLAKE2s() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "blake2s"
	p.Aliases = []string{"b2s"}
	p.Category = "hashs"
	p.Description = blake2Intro + "\n\nBLAKE2s is optimized for 8- to 32-bit platforms and produces\ndigests of any size between 1 and 32 bytes."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("key", "", "MAC key")
		flags.Int("len", 32, "length of the output hash in bytes, must be either 16 or 32")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		length := helpers.IntFlag(flags, "len", 32)
		if length != 16 && length != 32 {
			return errors.New("invalid length")
		}
		return doBLAKE2s(r, w, macKeyFlag(flags), length)
	}
	return p
}

func doBLAKE2b(r io.Reader, w io.Writer, macKey []byte, length int) error {
	hasher, err := blake2b.New(length, macKey)
	if err != nil {
		return err
	}
	if _, err := io.Copy(hasher, r); err != nil {
		return err
	}
	_, err = io.WriteString(w, fmt.Sprintf("%x", hasher.Sum(nil)))
	return err
}

// NewPluginBLAKE2b creates a plugin
func NewPluginBLAKE2b() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "blake2b"
	p.Aliases = []string{"b2b"}
	p.Category = "hashs"
	p.Description = blake2Intro + "\n\nBLAKE2b is optimized for 64-bit platforms and produces digests of\nany size between 1 and 64 bytes. This plugin defaults to 512-bit hashes."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("key", "", "MAC key")
		flags.Int("len", 64, "length of the output hash in bytes, must be between 1 and 64")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		length := helpers.IntFlag(flags, "len", 64)
		if length < 1 || length > 64 {
			length = 32
		}
		return doBLAKE2b(r, w, macKeyFlag(flags), length)
	}
	return p
}

func doBLAKE3(r io.Reader, w io.Writer, outLen int, key []byte, derive bool, ctx string) error {
	if len(key) > 0 && derive {
		derivedKey := make([]byte, len(key))
		blake3.DeriveKey(derivedKey, ctx, key)
		return writeHex(w, derivedKey)
	}
	var hasher hash.Hash
	if len(key) > 0 {
		hasher = blake3.New(outLen, key)
	} else {
		hasher = blake3.New(outLen, nil)
	}
	if _, err := io.Copy(hasher, r); err != nil {
		return err
	}
	return writeHex(w, hasher.Sum(nil))
}

// NewPluginBLAKE3 creates a plugin
func NewPluginBLAKE3() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "blake3"
	p.Aliases = []string{"b3"}
	p.Category = "hashs"
	p.Description = "BLAKE3 is a cryptographic hash function that is fast and secure.\nIt can be used as PRF, MAC, KDF, and XOF as well as a regular hash."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("key", "", "key (requires 32 bytes)")
		flags.Int("length", 32, "number of output bytes")
		flags.String("derive-key", "", "derive key")
		flags.String("context", "", "context for key derivation")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		outLen := helpers.IntFlag(flags, "length", 32)
		if outLen != 32 && outLen != 64 {
			outLen = 32
		}
		if deriveKey := helpers.StringFlag(flags, "derive-key"); deriveKey != "" {
			ctx := helpers.StringFlag(flags, "context")
			if ctx == "" {
				ctx = "BLAKE3 2020-02-13 13:33:37 test data context"
			}
			return doBLAKE3(r, w, outLen, []byte(deriveKey), true, ctx)
		}
		if rawKey := helpers.StringFlag(flags, "key"); rawKey != "" {
			if len(rawKey) != 32 {
				return errors.New("invalid key length")
			}
			return doBLAKE3(r, w, outLen, []byte(rawKey), false, "")
		}
		return doBLAKE3(r, w, outLen, nil, false, "")
	}
	return p
}
