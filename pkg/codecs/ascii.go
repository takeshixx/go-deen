package codecs

import (
	"flag"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func writeASCII(w io.Writer, data []byte, mode string) error {
	if mode == "" {
		mode = "strict"
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("input is not valid UTF-8")
	}
	for _, r := range string(data) {
		if r <= 0x7f {
			if _, err := fmt.Fprintf(w, "%c", r); err != nil {
				return err
			}
			continue
		}
		switch mode {
		case "strict":
			return fmt.Errorf("non-ASCII code point U+%04X", r)
		case "replace":
			if _, err := io.WriteString(w, "?"); err != nil {
				return err
			}
		case "strip":
			continue
		case "escape":
			if r <= 0xffff {
				if _, err := fmt.Fprintf(w, "\\u%04X", r); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "\\U%08X", r); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported ASCII mode %q", mode)
		}
	}
	return nil
}

// NewPluginASCII creates a UTF-8 to ASCII conversion plugin.
func NewPluginASCII() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "ascii"
	p.Aliases = []string{"to-ascii"}
	p.Category = "codecs"
	p.Description = "Convert UTF-8 text to ASCII using strict, replace, strip or escape handling for non-ASCII code points."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("mode", "strict", "non-ASCII handling mode (strict, replace, strip, escape)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return writeASCII(w, data, helpers.StringFlag(flags, "mode"))
	}
	return p
}
