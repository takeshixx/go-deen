package formatters

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/TylerBrock/colorjson"
	"github.com/itchyny/gojq"
	"github.com/takeshixx/deen/pkg/types"
)

func processJQ(task *types.DeenTask, queryStr string, plain, noColor bool) {
	go func() {
		defer task.Close()
		query, err := gojq.Parse(queryStr)
		if err != nil {
			log.Fatalln(err)
		}

		data, err := ioutil.ReadAll(task.Reader)
		if err != nil {
			task.ErrChan <- err
		}

		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			task.ErrChan <- err
		}

		iter := query.Run(jsonData)
		for {
			v, ok := iter.Next()
			if !ok {
				// TODO: send to task.ErrChan?
				break
			}
			if err, ok := v.(error); ok {
				task.ErrChan <- err
			}

			if !plain {
				if noColor {
					encoder := json.NewEncoder(task.PipeWriter)
					encoder.SetIndent("", "    ")
					err = encoder.Encode(v)
					if err != nil {
						task.ErrChan <- err
					}
				} else {
					f := colorjson.NewFormatter()
					f.Indent = 4
					colored, err := f.Marshal(v)
					if err != nil {
						task.ErrChan <- err
					}
					br := bytes.NewReader(colored)
					_, err = io.Copy(task.PipeWriter, br)
					if err != nil {
						task.ErrChan <- err
					}
				}
			} else {
				plainEncoder := json.NewEncoder(task.PipeWriter)
				err = plainEncoder.Encode(v)
				if err != nil {
					task.ErrChan <- err
				}
			}
		}
	}()
}

// NewPluginJQFormatter creates a new PluginJSONFormatter object
func NewPluginJQFormatter() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "jq"
	p.Aliases = []string{}
	p.Category = "formatters"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processJQ(task, "", false, true)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		queryFlag := flags.Lookup("q")
		queryStr := queryFlag.Value.String()
		if queryStr == "" {
			err := fmt.Errorf("No query provided")
			go func() {
				task.ErrChan <- err
				task.Close()
			}()
			return
		}
		plainFlag := flags.Lookup("plain")
		plain, err := strconv.ParseBool(plainFlag.Value.String())
		if err != nil {
			plain = true
		}
		noColorFlag := flags.Lookup("no-color")
		noColor, err := strconv.ParseBool(noColorFlag.Value.String())
		if err != nil {
			noColor = false
		}
		processJQ(task, queryStr, plain, noColor)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "JSON query plugin, similar to jq.\n\n")
			flags.PrintDefaults()
		}
		flags.String("q", "", "query string")
		flags.Bool("no-color", false, "omit colors in formatted output")
		flags.Bool("plain", false, "print unformatted token")
		flags.Parse(args)
		return flags
	}
	return
}
