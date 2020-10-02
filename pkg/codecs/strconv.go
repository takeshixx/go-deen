package codecs

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginStrconv creates a new PluginStrconv object
func NewPluginStrconv() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "strconv"
	p.Aliases = []string{".strconv", "str", ".str"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			str, err := ioutil.ReadAll(task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Failed to read input data")
			}
			quotedStr := strconv.QuoteToASCII(string(str))
			quotedStr = strings.TrimPrefix(quotedStr, "\"")
			quotedStr = strings.TrimSuffix(quotedStr, "\"")
			strReader := strings.NewReader(quotedStr)
			_, err = io.Copy(task.PipeWriter, strReader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into encoder in strconv failed")
			}
		}()
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.ProcessDeenTaskFunc(task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			str, err := ioutil.ReadAll(task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Failed to read input data")
			}
			strStr := string(str)
			strStr = fmt.Sprintf("\"%s\"", strStr)
			unquotedStr, err := strconv.Unquote(strStr)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Failed to unquote input data")
			}
			strReader := strings.NewReader(unquotedStr)
			_, err = io.Copy(task.PipeWriter, strReader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in Hex failed")
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
			fmt.Fprintf(os.Stderr, "Quote/Unquote strings and apply/remove escape characters.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
