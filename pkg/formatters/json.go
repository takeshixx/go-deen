package formatters

import (
	"bytes"
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

func processJSONFormatColored(task *types.DeenTask) {
	go func() {
		defer task.Close()
		data := make(map[string](interface{}))
		decoder := json.NewDecoder(task.Reader)
		decoder.Decode(&data)

		f := colorjson.NewFormatter()
		f.Indent = 4
		outBuf, err := f.Marshal(data)
		if err != nil {
			task.ErrChan <- err
		}
		writeBuf := bytes.NewReader(outBuf)
		_, err = io.Copy(task.PipeWriter, writeBuf)
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

func processJSONFormat(task *types.DeenTask) {
	go func() {
		defer task.Close()
		data := make(map[string](interface{}))
		decoder := json.NewDecoder(task.Reader)
		err := decoder.Decode(&data)
		if err != nil {
			task.ErrChan <- err
		}
		encoder := json.NewEncoder(task.PipeWriter)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(data)
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginJSONFormatter creates a new PluginJSONFormatter object
func NewPluginJSONFormatter() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "json"
	p.Aliases = []string{"json-format"}
	p.Type = "formatter"

	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processJSONFormat(task)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		noColorFlag := flags.Lookup("no-color")
		noColor, err := strconv.ParseBool(noColorFlag.Value.String())
		if err != nil {
			err = errors.New("Failed to parse --no-color option")
		}
		if noColor {
			processJSONFormat(task)
			return
		}
		processJSONFormatColored(task)
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
