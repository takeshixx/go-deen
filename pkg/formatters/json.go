package formatters

import (
	"encoding/json"
	"flag"
	"io"
	"regexp"

	"github.com/TylerBrock/colorjson"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"github.com/tdewolff/minify/v2"
	minijson "github.com/tdewolff/minify/v2/json"
)

// NewPluginJSONFormatter creates a new JSON formatter plugin.
func NewPluginJSONFormatter() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "json"
	p.Aliases = []string{".json"}
	p.Category = "formatters"
	p.Description = "Prettify JSON when processing and minify it when unprocessing."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("no-color", false, "omit colors in output")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		var data interface{}
		if err := json.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		if helpers.IsBoolFlag(flags, "no-color") {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "    ")
			return enc.Encode(data)
		}
		f := colorjson.NewFormatter()
		f.Indent = 4
		out, err := f.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		m := minify.New()
		m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), minijson.Minify)
		return m.Minify("text/json", w, r)
	}
	return p
}
