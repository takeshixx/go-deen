package formatters

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/TylerBrock/colorjson"
	"github.com/takeshixx/deen/pkg/types"
)

func processJSONFormat(reader io.Reader) (outBuf []byte, err error) {
	data := make(map[string](interface{}))
	decoder := json.NewDecoder(reader)
	decoder.Decode(&data)

	f := colorjson.NewFormatter()
	f.Indent = 4
	outBuf, err = f.Marshal(data)

	return
}

// NewPluginJSONFormatter creates a new PluginJSONFormatter object
func NewPluginJSONFormatter() (p types.DeenPlugin) {
	p.Name = "json"
	p.Aliases = []string{"json-format"}
	p.Type = "formatter"
	p.ProcessStreamFunc = func(reader io.Reader) (outBuf []byte, err error) {
		return processJSONFormat(reader)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		return processJSONFormat(reader)
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		jsonCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		jsonCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("JSON beautifier with colorized output.\n\n")
			jsonCmd.PrintDefaults()
		}
		jsonCmd.Parse(args)
		return jsonCmd
	}
	return
}
