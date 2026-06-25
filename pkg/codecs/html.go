package codecs

import (
	"flag"
	"html"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHTML creates a new HTML entity escaping plugin.
func NewPluginHTML() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "html"
	p.Aliases = []string{".html"}
	p.Category = "codecs"
	p.Description = "Escapes special characters like \"<\" to become \"&lt;\". It escapes\nonly five such characters: <, >, &, ' and \"."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, html.EscapeString(string(data)))
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, html.UnescapeString(string(data)))
		return err
	}
	return p
}
