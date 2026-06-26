package codecs

import (
	"encoding/base32"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func base32Encoding(flags *flag.FlagSet) *base32.Encoding {
	enc := base32.StdEncoding
	if helpers.IsBoolFlag(flags, "hex") {
		enc = base32.HexEncoding
	}
	if helpers.IsBoolFlag(flags, "no-pad") {
		enc = enc.WithPadding(base32.NoPadding)
	}
	return enc
}

// NewPluginBase32 creates a new base32 plugin (RFC 4648).
func NewPluginBase32() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "base32"
	p.Aliases = []string{".base32", "b32", ".b32"}
	p.Category = "codecs"
	p.Description = "Base32 encoding as specified by RFC 4648."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("hex", false, "use \"Extended Hex Alphabet\" defined in RFC 4648")
		flags.Bool("no-pad", false, "disable padding")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := base32Encoding(flags)
		return encodeStream(r, w, func(w io.Writer) io.WriteCloser { return base32.NewEncoder(enc, w) })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := base32Encoding(flags)
		return decodeTrimmed(r, w, func(r io.Reader) io.Reader { return base32.NewDecoder(enc, r) })
	}
	return p
}
