package formatters

import (
	"encoding/json"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
	"github.com/vmihailenco/msgpack/v5"
)

// NewPluginMessagePack creates a JSON-to-MessagePack encoder and decoder.
func NewPluginMessagePack() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "msgpack"
	p.Aliases = []string{"messagepack"}
	p.Category = "formatters"
	p.Description = "Encode JSON to MessagePack and decode MessagePack back to formatted JSON."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data interface{}
		if err := json.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		out, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var data interface{}
		if err := msgpack.Unmarshal(input, &data); err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(data)
	}
	return p
}
