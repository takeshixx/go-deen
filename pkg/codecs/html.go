package codecs

import (
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHTML creates a new PluginHTML object
func NewPluginHTML() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "html"
	p.Aliases = []string{".html"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf []byte
		var err error
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return outBuf, err
		}
		outBuf = []byte(html.EscapeString(string(data)))
		return outBuf, err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf []byte
		var err error
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return outBuf, err
		}
		escapedData := html.UnescapeString(string(data))
		outBuf = []byte(escapedData)
		return outBuf, err
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Escapes special characters like \"<\" to become \"&lt;\".\nIt escapes only five such characters: <, >, &, ' and \".\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
