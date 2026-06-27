package formatters

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginCBOR creates a JSON-to-CBOR encoder and decoder.
func NewPluginCBOR() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "cbor"
	p.Category = "formatters"
	p.Description = "Encode JSON to CBOR and decode CBOR back to formatted JSON."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		var data interface{}
		if err := json.NewDecoder(r).Decode(&data); err != nil {
			return err
		}
		out, err := cbor.Marshal(data)
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
		if err := cbor.Unmarshal(input, &data); err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(jsonable(data))
	}
	return p
}

func jsonable(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, v := range x {
			m[fmt.Sprint(k)] = jsonable(v)
		}
		return m
	case map[string]interface{}:
		for k, v := range x {
			x[k] = jsonable(v)
		}
		return x
	case []interface{}:
		for i, v := range x {
			x[i] = jsonable(v)
		}
		return x
	default:
		return v
	}
}
