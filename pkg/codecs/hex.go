package codecs

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginHex creates a new PluginHex object
func NewPluginHex() (p types.DeenPlugin) {
	p.Name = "hex"
	p.Aliases = []string{".hex", "asciihex", ".asciihex"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			encoder := hex.NewEncoder(task.PipeWriter)
			_, err := io.Copy(encoder, task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into encoder in Hex failed")
			}
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in Hex failed")
			}
		}()
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			wrappedReader := types.TrimReader{}
			wrappedReader.Rd = task.Reader
			decoder := hex.NewDecoder(wrappedReader)
			_, err := io.Copy(task.PipeWriter, decoder)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in Hex failed")
			}
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in Hex failed")
			}
		}()
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Apply ASCII hex encoding or decoding to data.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
