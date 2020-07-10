package codecs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func processBase64(encoding *base64.Encoding, reader io.Reader) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	encoder := base64.NewEncoder(encoding, &outBuf)
	if _, err := io.Copy(encoder, reader); err != nil {
		return outBuf.Bytes(), err
	}
	encoder.Close()
	return outBuf.Bytes(), err
}

func unprocessBase64(encoding *base64.Encoding, reader io.Reader) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	// We have to remove leading/trailing whitespaces
	wrappedReader := trimReader{}
	wrappedReader.rd = reader
	decoder := base64.NewDecoder(encoding, reader)
	wrapper := struct{ io.Writer }{&outBuf}
	if _, err := io.Copy(wrapper, decoder); err != nil {
		return outBuf.Bytes(), err
	}
	return outBuf.Bytes(), err
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

// NewPluginBase64 creates a new PluginBase64 object
func NewPluginBase64() (p types.DeenPlugin) {
	p.Name = "base64"
	p.Aliases = []string{".base64", "b64", ".b64"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return processBase64(base64.StdEncoding, reader)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		if raw, err := isRaw(flags); raw && err == nil {
			return processBase64(base64.RawStdEncoding, reader)
		}
		return processBase64(base64.StdEncoding, reader)
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var processed []byte
		var err error
		processed, err = unprocessBase64(base64.RawURLEncoding, reader)
		if err == nil {
			return processed, err
		}
		processed, err = unprocessBase64(base64.RawStdEncoding, reader)
		if err == nil {
			return processed, err
		}
		err = errors.New("Invalid Base64 data")
		return processed, err
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		if strict, err := isStrict(flags); strict && err == nil {
			if raw, err := isRaw(flags); raw && err == nil {
				return unprocessBase64(base64.RawStdEncoding, reader)
			}
			return unprocessBase64(base64.StdEncoding, reader)
		}
		var processed []byte
		var err error
		processed, err = unprocessBase64(base64.RawURLEncoding, reader)
		if err == nil {
			return processed, err
		}
		processed, err = unprocessBase64(base64.RawStdEncoding, reader)
		if err == nil {
			return processed, err
		}
		err = errors.New("Invalid Base64 data")
		return processed, err
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
		b64Cmd.Parse(args)
		return b64Cmd
	}
	return
}

// NewPluginBase64Url creates a new PluginBase64Url object
func NewPluginBase64Url() (p types.DeenPlugin) {
	p.Name = "base64url"
	p.Aliases = []string{".base64url", "b64u", ".b64u", "b64url", ".b64url"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return processBase64(base64.URLEncoding, reader)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		if raw, err := isRaw(flags); raw && err == nil {
			return processBase64(base64.RawURLEncoding, reader)
		}
		return processBase64(base64.URLEncoding, reader)
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return unprocessBase64(base64.URLEncoding, reader)
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		if raw, err := isRaw(flags); raw && err == nil {
			return unprocessBase64(base64.RawURLEncoding, reader)
		}
		return unprocessBase64(base64.URLEncoding, reader)
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		b64Cmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		b64Cmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Alternate Base64 encoding defined in RFC 4648 (typically used in URLs and file names).\n\n")
			b64Cmd.PrintDefaults()
		}
		b64Cmd.Bool("raw", false, "unpadded alternate Base64 encoding (as defined in RFC 4648 section 3.2)")
		b64Cmd.Parse(args)
		return b64Cmd
	}
	return
}
