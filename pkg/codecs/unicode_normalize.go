package codecs

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/text/unicode/norm"
)

func normalizationForm(command string, flags *flag.FlagSet) (norm.Form, error) {
	form := command
	if form != "nfc" && form != "nfd" && form != "nfkc" && form != "nfkd" {
		form = helpers.StringFlag(flags, "form")
	}
	switch strings.ToLower(form) {
	case "", "nfc":
		return norm.NFC, nil
	case "nfd":
		return norm.NFD, nil
	case "nfkc":
		return norm.NFKC, nil
	case "nfkd":
		return norm.NFKD, nil
	default:
		return norm.NFC, fmt.Errorf("unsupported normalization form %q", form)
	}
}

// NewPluginUnicodeNormalize creates a Unicode normalization plugin.
func NewPluginUnicodeNormalize() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "unicode-normalize"
	p.Aliases = []string{"normalize", "nfc", "nfd", "nfkc", "nfkd"}
	p.Category = "codecs"
	p.Description = "Normalize UTF-8 text to NFC, NFD, NFKC or NFKD so visually similar strings have a consistent representation."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("form", "nfc", "normalization form (nfc, nfd, nfkc, nfkd)")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		form, err := normalizationForm(p.Command, flags)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, form.Reader(r))
		return err
	}
	return p
}
