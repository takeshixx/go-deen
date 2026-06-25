package codecs

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginPEM creates a new PEM (RFC 1421) plugin.
func NewPluginPEM() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "pem"
	p.Aliases = []string{".pem"}
	p.Category = "codecs"
	p.Description = "Privacy Enhanced Mail (PEM) data encoding/decoding (RFC 1421)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("type", "MESSAGE", "data type")
		flags.String("headers", "", "message headers in JSON format")
		flags.Bool("cert", false, "create a PEM encoded certificate")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		headers := map[string]string{}
		if raw := helpers.StringFlag(flags, "headers"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &headers); err != nil {
				return err
			}
		}
		dataType := helpers.StringFlag(flags, "type")
		if dataType == "" {
			dataType = "MESSAGE"
		}
		if helpers.IsBoolFlag(flags, "cert") {
			dataType = "CERTIFICATE"
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return pem.Encode(w, &pem.Block{Type: dataType, Headers: headers, Bytes: data})
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(data)
		if block == nil {
			return errors.New("no PEM block found in input")
		}
		_, err = w.Write(block.Bytes)
		return err
	}
	return p
}
