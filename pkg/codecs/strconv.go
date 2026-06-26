package codecs

import (
	"flag"
	"io"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginStrconv creates a new strconv quoting plugin.
func NewPluginStrconv() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "strconv"
	p.Aliases = []string{".strconv", "str", ".str"}
	p.Category = "codecs"
	p.Description = "Quote/Unquote strings and apply/remove escape characters."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("ctrl", false, "only escape control sequences")
		flags.Bool("graph", false, "escape to graphs")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var quoted string
		switch {
		case helpers.IsBoolFlag(flags, "ctrl"):
			quoted = strconv.Quote(string(data))
		case helpers.IsBoolFlag(flags, "graph"):
			quoted = strconv.QuoteToGraphic(string(data))
		default:
			quoted = strconv.QuoteToASCII(string(data))
		}
		quoted = strings.TrimPrefix(quoted, "\"")
		quoted = strings.TrimSuffix(quoted, "\"")
		_, err = io.WriteString(w, quoted)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		unquoted, err := strconv.Unquote("\"" + string(data) + "\"")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, unquoted)
		return err
	}
	return p
}
