package formatters

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
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
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "JSON beautifier with colorized output.\n\n")
			flags.PrintDefaults()
		}
		flags.Bool("no-color", false, "omit colors in output")
		flags.Parse(args)
		return flags
	}
	return
}
