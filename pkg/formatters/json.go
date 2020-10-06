package formatters

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/TylerBrock/colorjson"
	"github.com/takeshixx/deen/pkg/types"
	"github.com/tdewolff/minify/v2"
	minijson "github.com/tdewolff/minify/v2/json"
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
	p.Name = "jsons"
	p.Aliases = []string{".json", "json-format"}
	p.Category = "formatter"
	p.Unprocess = false
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

	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			minifier := minify.New()
			minifier.AddFuncRegexp(regexp.MustCompile("[/+]json$"), minijson.Minify)
			err := minifier.Minify("text/json", task.PipeWriter, task.Reader)
			if err != nil {
				task.ErrChan <- err
			}
		}()
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}

	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "JSON formatter plugin that processes JSON to a readable,\nprettified representation, and unprocesses beautified\nJSON to minified JSON.\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			flags.Bool("no-color", false, "omit colors in output")
		}
		flags.Parse(args)
		return flags
	}
	return
}
