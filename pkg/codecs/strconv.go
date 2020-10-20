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

func processStrconv(t *types.DeenTask, ctrlOnly bool) {
	go func() {
		defer t.Close()
		str, err := ioutil.ReadAll(t.Reader)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Failed to read input data")
		}
		var quotedStr string
		if ctrlOnly {
			quotedStr = strconv.Quote(string(str))
		} else {
			quotedStr = strconv.QuoteToASCII(string(str))
		}
		quotedStr = strings.TrimPrefix(quotedStr, "\"")
		quotedStr = strings.TrimSuffix(quotedStr, "\"")
		strReader := strings.NewReader(quotedStr)
		_, err = io.Copy(t.PipeWriter, strReader)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Copying into encoder in strconv failed")
		}
	}()
}

// NewPluginStrconv creates a new PluginStrconv object
func NewPluginStrconv() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "strconv"
	p.Aliases = []string{".strconv", "str", ".str"}
	p.Category = "codecs"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processStrconv(task, false)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		ctrlOnlyPtr := flags.Lookup("ctrl")
		ctflOnly := false
		ctflOnly, _ = strconv.ParseBool(ctrlOnlyPtr.Value.String())
		processStrconv(task, ctflOnly)
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
		if !self.Unprocess {
			flags.Bool("ctrl", false, "only escape control sequences")
		}
		flags.Parse(args)
		return flags
	}
	return
}
