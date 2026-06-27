package codecs

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/takeshixx/deen/pkg/types"
)

type unicodeInspection struct {
	bytes       int
	utf8Valid   bool
	runes       int
	invalidUTF8 int
	controls    int
	bom         string
	likely      string
	nullEven    int
	nullOdd     int
	nullMod4    [4]int
}

func inspectUnicode(data []byte) unicodeInspection {
	info := unicodeInspection{
		bytes:     len(data),
		utf8Valid: utf8.Valid(data),
		bom:       "none",
		likely:    "binary or unknown",
	}
	info.bom = unicodeBOM(data)
	info.likely = likelyUnicodeEncoding(data, info.bom, info.utf8Valid)
	for i, b := range data {
		if b == 0 {
			if i%2 == 0 {
				info.nullEven++
			} else {
				info.nullOdd++
			}
			info.nullMod4[i%4]++
		}
	}
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			info.invalidUTF8++
		}
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			info.controls++
		}
		info.runes++
		data = data[size:]
	}
	return info
}

func unicodeBOM(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}):
		return "UTF-8"
	case bytes.HasPrefix(data, []byte{0xff, 0xfe, 0x00, 0x00}):
		return "UTF-32LE"
	case bytes.HasPrefix(data, []byte{0x00, 0x00, 0xfe, 0xff}):
		return "UTF-32BE"
	case bytes.HasPrefix(data, []byte{0xff, 0xfe}):
		return "UTF-16LE"
	case bytes.HasPrefix(data, []byte{0xfe, 0xff}):
		return "UTF-16BE"
	default:
		return "none"
	}
}

func likelyUnicodeEncoding(data []byte, bom string, validUTF8 bool) string {
	if len(data) == 0 {
		return "empty"
	}
	if bom != "none" {
		return bom + " with BOM"
	}
	nulls := nullPattern(data)
	var evenNulls, oddNulls int
	for i, count := range nulls {
		if i%2 == 0 {
			evenNulls += count
		} else {
			oddNulls += count
		}
	}
	switch {
	case nulls[1] > 0 && nulls[2] > 0 && nulls[3] > 0 && nulls[0] == 0:
		return "likely UTF-32LE"
	case nulls[0] > 0 && nulls[1] > 0 && nulls[2] > 0 && nulls[3] == 0:
		return "likely UTF-32BE"
	case oddNulls > 0 && evenNulls == 0:
		return "likely UTF-16LE"
	case evenNulls > 0 && oddNulls == 0:
		return "likely UTF-16BE"
	}
	if validUTF8 {
		if asciiOnly(data) {
			return "ASCII-compatible UTF-8"
		}
		return "UTF-8"
	}
	return "binary or unknown"
}

func asciiOnly(data []byte) bool {
	for _, b := range data {
		if b >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func nullPattern(data []byte) [4]int {
	var counts [4]int
	for i, b := range data {
		if b == 0 {
			counts[i%4]++
		}
	}
	return counts
}

// NewPluginUnicodeInspect creates a Unicode and text encoding inspector.
func NewPluginUnicodeInspect() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "unicode-inspect"
	p.Aliases = []string{"utfinspect", "charset"}
	p.Category = "codecs"
	p.Description = "Inspect text bytes for UTF-8 validity, BOMs, code point counts and likely UTF-16/UTF-32 byte order."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		info := inspectUnicode(data)
		_, err = fmt.Fprintf(w, "bytes: %d\nlikely: %s\nbom: %s\nutf-8 valid: %t\ncode points: %d\ninvalid utf-8 bytes: %d\ncontrol code points: %d\nnull bytes: even=%d odd=%d mod4=[%d %d %d %d]\n",
			info.bytes,
			info.likely,
			info.bom,
			info.utf8Valid,
			info.runes,
			info.invalidUTF8,
			info.controls,
			info.nullEven,
			info.nullOdd,
			info.nullMod4[0],
			info.nullMod4[1],
			info.nullMod4[2],
			info.nullMod4[3],
		)
		return err
	}
	return p
}
