package formatters

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/TylerBrock/colorjson"
	"github.com/itchyny/gojq"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginJQFormatter creates a new jq-like JSON query plugin.
func NewPluginJQFormatter() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "jq"
	p.Category = "formatters"
	p.Description = "JSON query plugin, similar to jq."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("q", "", "query string")
		flags.Bool("no-color", false, "omit colors in formatted output")
		flags.Bool("plain", false, "print unformatted token")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		queryStr := helpers.StringFlag(flags, "q")
		if queryStr == "" {
			return fmt.Errorf("no query provided (use -q)")
		}
		query, err := gojq.Parse(queryStr)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return err
		}

		plain := helpers.IsBoolFlag(flags, "plain")
		noColor := helpers.IsBoolFlag(flags, "no-color")

		iter := query.Run(jsonData)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return err
			}
			switch {
			case plain:
				enc := json.NewEncoder(w)
				if err := enc.Encode(v); err != nil {
					return err
				}
			case noColor:
				enc := json.NewEncoder(w)
				enc.SetIndent("", "    ")
				if err := enc.Encode(v); err != nil {
					return err
				}
			default:
				f := colorjson.NewFormatter()
				f.Indent = 4
				colored, err := f.Marshal(v)
				if err != nil {
					return err
				}
				if _, err := w.Write(colored); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return p
}
