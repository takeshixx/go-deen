package formatters

import (
	"flag"
	"io"

	"github.com/clbanning/mxj/v2"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginJSON2XML creates a reversible JSON/XML conversion plugin.
// Processing converts JSON to XML; unprocessing (".json2xml") converts XML
// back to JSON.
func NewPluginJSON2XML() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "json2xml"
	p.Aliases = []string{".json2xml"}
	p.Category = "formatters"
	p.Description = "Convert JSON to XML. Use .json2xml to convert XML back to JSON."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		m, err := mxj.NewMapJson(data)
		if err != nil {
			return err
		}
		out, err := m.XmlIndent("", "    ")
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		m, err := mxj.NewMapXml(data)
		if err != nil {
			return err
		}
		out, err := m.JsonIndent("", "    ")
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	return p
}
