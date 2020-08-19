package codecs

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
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

func isRaw(flags *flag.FlagSet) (raw bool, err error) {
	rawFlag := flags.Lookup("raw")
	raw, err = strconv.ParseBool(rawFlag.Value.String())
	if err != nil {
		err = errors.Wrap(err, "Failed to parse --raw option")
		return
	}
	return
}

func isStrict(flags *flag.FlagSet) (raw bool, err error) {
	rawFlag := flags.Lookup("strict")
	raw, err = strconv.ParseBool(rawFlag.Value.String())
	if err != nil {
		err = errors.Wrap(err, "Failed to parse --strict option")
		return
	}
	return
}

func isURLSafe(flags *flag.FlagSet) (url bool, err error) {
	urlFlag := flags.Lookup("url")
	url, err = strconv.ParseBool(urlFlag.Value.String())
	if err != nil {
		err = errors.Wrap(err, "Failed to parse --url option")
		return
	}
	return
}

func parseBase64Encoding(flags *flag.FlagSet) (enc *base64.Encoding) {
	var raw, strict, url bool
	strict, _ = isStrict(flags)
	raw, _ = isRaw(flags)
	url, _ = isURLSafe(flags)
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
func NewPluginBase64() (p types.DeenPlugin) {
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
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		b64Cmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		b64Cmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Base64 encoding defined in RFC 4648. By default, decoding tries to decode raw URL and default Base64 data.\n\n")
			b64Cmd.PrintDefaults()
		}
		b64Cmd.Bool("strict", false, "use strict Base64 decoding mode (don't try different encodings)")
		b64Cmd.Bool("raw", false, "unpadded Base64 encoding (as defined in RFC 4648 section 3.2)")
		b64Cmd.Bool("url", false, "use alternate, URL-safe Base64 encoding")
		b64Cmd.Parse(args)
		return b64Cmd
	}
	return
}
