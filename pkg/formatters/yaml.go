package formatters

import (
	"encoding/json"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
	"gopkg.in/yaml.v3"
)

// NewPluginYAMLFormatter creates a YAML formatter and YAML-to-JSON converter.
func NewPluginYAMLFormatter() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "yaml"
	p.Aliases = []string{"yml", ".yaml", ".yml"}
	p.Category = "formatters"
	p.Description = "Normalize YAML when processing and convert YAML to compact JSON when unprocessing."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data interface{}
		if err := yaml.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		out, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data interface{}
		if err := yaml.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(data)
	}
	return p
}
