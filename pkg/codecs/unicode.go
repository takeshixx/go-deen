package codecs

import (
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

func readEndianness(flags *flag.FlagSet) unicode.Endianness {
	if helpers.IsBoolFlag(flags, "big") {
		return unicode.BigEndian
	}
	return unicode.LittleEndian
}

func readBOMPolicy(flags *flag.FlagSet) unicode.BOMPolicy {
	switch helpers.StringFlag(flags, "bom") {
	case "use":
		return unicode.UseBOM
	case "expect":
		return unicode.ExpectBOM
	default:
		return unicode.IgnoreBOM
	}
}

// unicodeEncoding maps a command and flags to a text encoding.
func unicodeEncoding(command string, flags *flag.FlagSet) encoding.Encoding {
	endianness := readEndianness(flags)
	bomPolicy := readBOMPolicy(flags)
	switch command {
	case "utf16":
		return unicode.UTF16(endianness, bomPolicy)
	case "utf32":
		return utf32.UTF32(utf32.Endianness(endianness), utf32.BOMPolicy(bomPolicy))
	case "euckr":
		return korean.EUCKR
	default:
		return unicode.UTF8
	}
}

// NewPluginUnicode creates a new unicode (re)encoding plugin.
func NewPluginUnicode() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "unicode"
	p.Aliases = []string{".unicode", "utf8", ".utf8", "utf16", ".utf16", "utf32", ".utf32", "euckr", ".euckr"}
	p.Category = "codecs"
	p.Description = "Encode/decode data between UTF-8 and UTF-16, UTF-32 or EUC-KR.\nThe target encoding is selected by the command alias (e.g. utf16)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("bom", "ignore", "BOM mode (use, ignore, expect)")
		flags.Bool("big", false, "use big endian (default: little)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := unicodeEncoding(p.Command, flags)
		_, err := io.Copy(enc.NewEncoder().Writer(w), r)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := unicodeEncoding(p.Command, flags)
		_, err := io.Copy(w, enc.NewDecoder().Reader(r))
		return err
	}
	return p
}
