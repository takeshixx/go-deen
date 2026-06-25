package codecs

import (
	"flag"
	"io"
	"net/url"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginURL creates a new URL query escaping plugin.
func NewPluginURL() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "url"
	p.Aliases = []string{".url", "urlencode", ".urlencode"}
	p.Category = "codecs"
	p.Description = "Escapes the string so it can be safely placed inside a URL query."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, url.QueryEscape(string(data)))
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		decoded, err := url.QueryUnescape(string(data))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, decoded)
		return err
	}
	return p
}
