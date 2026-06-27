package formatters

import (
	"encoding/json"
	"flag"
	"io"

	"github.com/BurntSushi/toml"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginTOML creates a TOML formatter and TOML-to-JSON converter.
func NewPluginTOML() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "toml"
	p.Category = "formatters"
	p.Description = "Normalize TOML when processing and convert TOML to JSON when unprocessing."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data map[string]interface{}
		if _, err := toml.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		return toml.NewEncoder(w).Encode(data)
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data map[string]interface{}
		if _, err := toml.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(data)
	}
	return p
}
