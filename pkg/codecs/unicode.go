package codecs

import (
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

func readEndianness(flags *flag.FlagSet) unicode.Endianness {
	if helpers.IsBoolFlag(flags, "big") {
		return unicode.BigEndian
	}
	return unicode.LittleEndian
}

func commandEndianness(command string, flags *flag.FlagSet) unicode.Endianness {
	switch command {
	case "utf16be", "utf32be":
		return unicode.BigEndian
	case "utf16le", "utf32le":
		return unicode.LittleEndian
	default:
		return readEndianness(flags)
	}
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
	if command == "" || command == "unicode" {
		command = helpers.StringFlag(flags, "encoding")
	}
	endianness := commandEndianness(command, flags)
	bomPolicy := readBOMPolicy(flags)
	switch command {
	case "utf16", "utf16le", "utf16be":
		return unicode.UTF16(endianness, bomPolicy)
	case "utf32", "utf32le", "utf32be":
		return utf32.UTF32(utf32.Endianness(endianness), utf32.BOMPolicy(bomPolicy))
	case "latin1", "iso88591", "iso-8859-1":
		return charmap.ISO8859_1
	case "windows1252", "windows-1252", "cp1252":
		return charmap.Windows1252
	case "shiftjis", "shift-jis", "sjis":
		return japanese.ShiftJIS
	case "eucjp", "euc-jp":
		return japanese.EUCJP
	case "gbk":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	case "euckr":
		return korean.EUCKR
	case "koi8r", "koi8-r":
		return charmap.KOI8R
	default:
		return unicode.UTF8
	}
}

// NewPluginUnicode creates a new unicode (re)encoding plugin.
func NewPluginUnicode() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "unicode"
	p.Aliases = []string{
		".unicode",
		"utf8", ".utf8",
		"utf16", ".utf16", "utf16le", ".utf16le", "utf16be", ".utf16be",
		"utf32", ".utf32", "utf32le", ".utf32le", "utf32be", ".utf32be",
		"latin1", ".latin1", "iso88591", ".iso88591", "iso-8859-1", ".iso-8859-1",
		"windows1252", ".windows1252", "windows-1252", ".windows-1252", "cp1252", ".cp1252",
		"shiftjis", ".shiftjis", "shift-jis", ".shift-jis", "sjis", ".sjis",
		"eucjp", ".eucjp", "euc-jp", ".euc-jp",
		"gbk", ".gbk", "gb18030", ".gb18030", "big5", ".big5",
		"euckr", ".euckr", "koi8r", ".koi8r", "koi8-r", ".koi8-r",
	}
	p.Category = "codecs"
	p.Description = "Encode/decode data between UTF-8, UTF-16, UTF-32 and common legacy character sets.\nThe target encoding is selected by the command alias (e.g. utf16le or windows1252).\nThis is a text transcoder: invalid sequences are replaced, so it is\nnot byte-preserving for arbitrary binary input."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("encoding", "utf8", "text encoding (utf8, utf16le, utf16be, utf32le, utf32be, latin1, windows1252, shiftjis, eucjp, gbk, gb18030, big5, euckr, koi8r)")
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
