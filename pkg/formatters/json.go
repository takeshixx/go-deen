package formatters

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/TylerBrock/colorjson"
	"github.com/takeshixx/deen/pkg/types"
)

func processJSONFormatColored(reader io.Reader) (outBuf []byte, err error) {
	data := make(map[string](interface{}))
	decoder := json.NewDecoder(reader)
	decoder.Decode(&data)

	f := colorjson.NewFormatter()
	f.Indent = 4
	outBuf, err = f.Marshal(data)

	return
}

func processJSONFormat(reader io.Reader) (outBuf []byte, err error) {
	data := make(map[string](interface{}))
	decoder := json.NewDecoder(reader)
	decoder.Decode(&data)

	outBuf, err = json.MarshalIndent(data, "", "")
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
		noColorFlag := flags.Lookup("no-color")
		noColor, err := strconv.ParseBool(noColorFlag.Value.String())
		if err != nil {
			err = errors.New("Failed to parse --no-color option")
			var outBuf []byte
			return outBuf, err
		}
		if noColor {
			return processJSONFormat(reader)
		} else {
			return processJSONFormatColored(reader)
		}
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		jsonCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		jsonCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("JSON beautifier with colorized output.\n\n")
			jsonCmd.PrintDefaults()
		}
		jsonCmd.Bool("no-color", false, "omit colors in output")
		jsonCmd.Parse(args)
		return jsonCmd
	}
	return
}
