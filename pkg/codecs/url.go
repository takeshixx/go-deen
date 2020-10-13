package codecs

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginURL creates a new PluginUrl object
func NewPluginURL() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "url"
	p.Aliases = []string{".url", "urlencode", ".urlencode"}
	p.Category = "codecs"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf []byte
		var err error
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return outBuf, err
		}
		outBuf = []byte(url.QueryEscape(string(data)))
		return outBuf, err
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf []byte
		var err error
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return outBuf, err
		}
		decodedData, err := url.QueryUnescape(string(data))
		if err != nil {
			return outBuf, err
		}
		outBuf = []byte(decodedData)
		return outBuf, err
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Escapes the string so it can be safely placed inside a URL query.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
