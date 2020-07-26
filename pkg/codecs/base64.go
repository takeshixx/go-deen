package codecs

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func processBase64Pipe(encoding *base64.Encoding, reader io.Reader, writer *io.PipeWriter) (err error) {
	go func() {
		encoder := base64.NewEncoder(encoding, writer)
		if _, err := io.Copy(encoder, reader); err != nil {
			return
		}
		encoder.Close()
		writer.Close()
	}()

	return
}

func unprocessBase64Pipe(encoding *base64.Encoding, reader io.Reader, writer *io.PipeWriter) (err error) {
	go func() {
		decoder := base64.NewDecoder(encoding, reader)
		if _, err := io.Copy(writer, decoder); err != nil {
			return
		}
		writer.Close()
	}()

	return
}

func isRaw(flags *flag.FlagSet) (raw bool, err error) {
	rawFlag := flags.Lookup("raw")
	raw, err = strconv.ParseBool(rawFlag.Value.String())
	if err != nil {
		err = errors.New("Failed to parse --raw option")
		return
	}
	return
}

func isStrict(flags *flag.FlagSet) (raw bool, err error) {
	rawFlag := flags.Lookup("strict")
	raw, err = strconv.ParseBool(rawFlag.Value.String())
	if err != nil {
		err = errors.New("Failed to parse --strict option")
		return
	}
	return
}

func isURLSafe(flags *flag.FlagSet) (url bool, err error) {
	urlFlag := flags.Lookup("url")
	url, err = strconv.ParseBool(urlFlag.Value.String())
	if err != nil {
		err = errors.New("Failed to parse --url option")
		return
	}
	return
}

// NewPluginBase64 creates a new PluginBase64 object
func NewPluginBase64() (p types.DeenPlugin) {
	p.Name = "base64"
	p.Aliases = []string{".base64", "b64", ".b64"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessPipeFunc = func(reader io.Reader, writer *io.PipeWriter) (err error) {
		return processBase64Pipe(base64.StdEncoding, reader, writer)
	}
	p.UnprocessPipeFunc = func(reader io.Reader, writer *io.PipeWriter) (err error) {
		err = unprocessBase64Pipe(base64.RawURLEncoding, reader, writer)
		if err == nil {
			return
		}
		err = unprocessBase64Pipe(base64.RawStdEncoding, reader, writer)
		if err == nil {
			return
		}
		err = errors.New("Invalid Base64 data")
		return
	}
	p.ProcessPipeWithFlags = func(flags *flag.FlagSet, reader io.Reader, writer *io.PipeWriter) (err error) {
		if url, err := isURLSafe(flags); url && err == nil {
			if raw, err := isRaw(flags); raw && err == nil {
				return processBase64Pipe(base64.RawURLEncoding, reader, writer)
			}
			return processBase64Pipe(base64.URLEncoding, reader, writer)
		}
		if raw, err := isRaw(flags); raw && err == nil {
			return processBase64Pipe(base64.RawStdEncoding, reader, writer)
		}
		return processBase64Pipe(base64.StdEncoding, reader, writer)
	}
	p.UnprocessPipeWithFlags = func(flags *flag.FlagSet, reader io.Reader, writer *io.PipeWriter) (err error) {
		if strict, err := isStrict(flags); strict && err == nil {
			if raw, err := isRaw(flags); raw && err == nil {
				return unprocessBase64Pipe(base64.RawStdEncoding, reader, writer)
			}
			return unprocessBase64Pipe(base64.StdEncoding, reader, writer)
		}
		err = unprocessBase64Pipe(base64.RawURLEncoding, reader, writer)
		if err == nil {
			return
		}
		err = unprocessBase64Pipe(base64.RawStdEncoding, reader, writer)
		if err == nil {
			return
		}
		err = errors.New("Invalid Base64 data")
		return
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
