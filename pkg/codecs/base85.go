package codecs

import (
	"encoding/ascii85"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginBase85 creates a new PluginBase85 object
func NewPluginBase85() (p types.DeenPlugin) {
	p.Name = "base85"
	p.Aliases = []string{".base85", "b85", ".b85",
		"ascii85", ".ascii85", "a85",
		".a85"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			encoder := ascii85.NewEncoder(task.PipeWriter)
			_, err := io.Copy(encoder, task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into encoder in Base85 failed")
			}
			err = encoder.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing encoder in Base85 failed")
			}
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in Base85 failed")
			}
		}()
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			wrappedReader := types.TrimReader{}
			wrappedReader.Rd = task.Reader
			decoder := ascii85.NewDecoder(wrappedReader)
			_, err := io.Copy(task.PipeWriter, decoder)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in Base85 failed")
			}
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in Base85 failed")
			}
		}()
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Implements the ascii85 data encoding as used in\nthe btoa tool and Adobe's PostScript and PDF\ndocument formats.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
