package codecs

import (
	"encoding/base32"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func processBase32(encoding *base32.Encoding, task *types.DeenTask) {
	go func() {
		encoder := base32.NewEncoder(encoding, task.PipeWriter)
		_, err := io.Copy(encoder, task.Reader)
		if err != nil {
			task.ErrChan <- err
		}
		err = encoder.Close()
		if err != nil {
			task.ErrChan <- err
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

func unprocessBas32(encoding *base32.Encoding, task *types.DeenTask) {
	go func() {
		decoder := base32.NewDecoder(encoding, task.Reader)
		_, err := io.Copy(task.PipeWriter, decoder)
		if err != nil {
			task.ErrChan <- err
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginBase32 creates a new PluginBase32 object
// Standard base32 encoding, as defined in RFC 4648
func NewPluginBase32() (p types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "base32"
	p.Aliases = []string{".base32", "b32", ".b32"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processBase32(base32.StdEncoding, task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		unprocessBas32(base32.StdEncoding, task)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		var enc *base32.Encoding
		noPad := helpers.IsBoolFlag(flags, "no-pad")
		if helpers.IsBoolFlag(flags, "hex") {
			enc = base32.HexEncoding
		} else {
			enc = base32.StdEncoding
		}
		if noPad {
			enc = enc.WithPadding(base32.NoPadding)
		}
		processBase32(enc, task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		var enc *base32.Encoding
		if helpers.IsBoolFlag(flags, "hex") {
			enc = base32.HexEncoding
		} else {
			enc = base32.StdEncoding
		}
		unprocessBas32(enc, task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Base32 encoding as specified by RFC 4648.\n\n")
			flags.PrintDefaults()
		}
		flags.Bool("hex", false, "use \"Extended Hex Alphabet\" defined in RFC 4648")
		if !self.Unprocess {
			flags.Bool("no-pad", false, "disable padding")
		}
		flags.Parse(args)
		return flags
	}
	return
}
