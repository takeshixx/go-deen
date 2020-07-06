package codecs

import (
	"html"
	"io"
	"io/ioutil"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHTML creates a new PluginHTML object
func NewPluginHTML() (p types.DeenPlugin) {
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
	return
}
