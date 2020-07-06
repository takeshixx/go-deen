package codecs

import (
	"io"
	"io/ioutil"
	"net/url"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginURL creates a new PluginUrl object
func NewPluginURL() (p types.DeenPlugin) {
	p.Name = "url"
	p.Aliases = []string{".url", "urlencode", ".urlencode"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf []byte
		var err error
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			//log.Fatal(err)
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
	return
}
