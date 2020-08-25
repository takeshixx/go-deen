package codecs

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func processBase64(encoding *base64.Encoding, task *types.DeenTask) {
	go func() {
		encoder := base64.NewEncoder(encoding, task.PipeWriter)
		_, err := io.Copy(encoder, task.Reader)
		if err != nil {
			task.ErrChan <- errors.Wrap(err, "Copying into encoder in processBase64 failed")
		}
		err = encoder.Close()
		if err != nil {
			task.ErrChan <- errors.Wrap(err, "Closing encoder in processBase64 failed")
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in processBase64 failed")
		}
	}()
}

func unprocessBase64(encoding *base64.Encoding, task *types.DeenTask) {
	go func() {
		decoder := base64.NewDecoder(encoding, task.Reader)
		_, err := io.Copy(task.PipeWriter, decoder)
		if err != nil {
			task.ErrChan <- errors.Wrap(err, "Copy in unprocessBase64 failed")
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in unprocessBase64 failed")
		}
	}()
}

func parseBase64Encoding(flags *flag.FlagSet) (enc *base64.Encoding) {
	raw := helpers.IsBoolFlag(flags, "raw")
	url := helpers.IsBoolFlag(flags, "url")
	strict := helpers.IsBoolFlag(flags, "strict")
	if strict {
		enc = base64.StdEncoding
	} else {
		if url && raw {
			enc = base64.RawURLEncoding
		} else if url {
			enc = base64.URLEncoding
		} else if raw {
			enc = base64.RawStdEncoding
		} else {
			enc = base64.StdEncoding
		}
	}
	return
}

// NewPluginBase64 creates a new PluginBase64 object
func NewPluginBase64() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "base64"
	p.Aliases = []string{".base64", "b64", ".b64"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processBase64(base64.StdEncoding, task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		unprocessBase64(base64.StdEncoding, task)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		enc := parseBase64Encoding(flags)
		processBase64(enc, task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		enc := parseBase64Encoding(flags)
		unprocessBase64(enc, task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Base64 encoding defined in RFC 4648. By default, decoding\ntries to decode raw URL and default Base64 data.\n\n")
			flags.PrintDefaults()
		}
		flags.Bool("strict", false, "use strict Base64 decoding mode (don't try different encodings)")
		flags.Bool("raw", false, "unpadded Base64 encoding (as defined in RFC 4648 section 3.2)")
		flags.Bool("url", false, "use alternate, URL-safe Base64 encoding")
		flags.Parse(args)
		return flags
	}
	return
}
