package formatters

import (
	"bytes"
	"encoding/xml"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// reformatXML re-encodes the XML read from r, dropping formatting-only
// whitespace text nodes. When indent is non-empty the output is pretty-printed,
// otherwise it is minified.
func reformatXML(r io.Reader, w io.Writer, indent string) error {
	dec := xml.NewDecoder(r)
	enc := xml.NewEncoder(w)
	if indent != "" {
		enc.Indent("", indent)
	}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Skip whitespace-only character data so re-indentation is clean.
		if cd, ok := tok.(xml.CharData); ok && len(bytes.TrimSpace(cd)) == 0 {
			continue
		}
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
	}
	return enc.Flush()
}

// NewPluginXMLFormatter creates an XML prettify/minify plugin.
func NewPluginXMLFormatter() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "xml"
	p.Aliases = []string{".xml"}
	p.Category = "formatters"
	p.Description = "Prettify XML when processing and minify it when unprocessing."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return reformatXML(r, w, "    ")
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return reformatXML(r, w, "")
	}
	return p
}
