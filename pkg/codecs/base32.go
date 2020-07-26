package codecs

import (
	"encoding/base32"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func processBase32Pipe(encoding *base32.Encoding, reader io.Reader, writer *io.PipeWriter) (err error) {
	go func() {
		encoder := base32.NewEncoder(encoding, writer)
		if _, err := io.Copy(encoder, reader); err != nil {
			return
		}
		encoder.Close()
		writer.Close()
	}()

	return
}

func unprocessBas32Pipe(encoding *base32.Encoding, reader io.Reader, writer *io.PipeWriter) (err error) {
	go func() {
		decoder := base32.NewDecoder(encoding, reader)
		if _, err := io.Copy(writer, decoder); err != nil {
			return
		}
		writer.Close()
	}()

	return
}

func isHex(flags *flag.FlagSet) (hex bool, err error) {
	hexFlag := flags.Lookup("hex")
	hex, err = strconv.ParseBool(hexFlag.Value.String())
	if err != nil {
		err = errors.New("Failed to parse --hex option")
		return
	}
	return
}

func isNoPad(flags *flag.FlagSet) (noPad bool, err error) {
	noPadFlag := flags.Lookup("no-pad")
	noPad, err = strconv.ParseBool(noPadFlag.Value.String())
	if err != nil {
		err = errors.New("Failed to parse --no-pad option")
		return
	}
	return
}

// NewPluginBase32 creates a new PluginBase32 object
// Standard base32 encoding, as defined in RFC 4648
func NewPluginBase32() (p types.DeenPlugin) {
	p.Name = "base32"
	p.Aliases = []string{".base32", "b32", ".b32"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessPipeFunc = func(reader io.Reader, writer *io.PipeWriter) (err error) {
		return processBase32Pipe(base32.StdEncoding, reader, writer)
	}
	p.UnprocessPipeFunc = func(reader io.Reader, writer *io.PipeWriter) (err error) {
		err = unprocessBas32Pipe(base32.StdEncoding, reader, writer)
		if err == nil {
			return
		}
		err = unprocessBas32Pipe(base32.HexEncoding, reader, writer)
		if err == nil {
			return
		}
		err = errors.New("Invalid Base32 data")
		return
	}
	p.ProcessPipeWithFlags = func(flags *flag.FlagSet, reader io.Reader, writer *io.PipeWriter) (err error) {
		var enc *base32.Encoding
		noPad, err := isNoPad(flags)
		if err != nil {
			return
		}
		if hex, err := isHex(flags); hex && err == nil {
			enc = base32.HexEncoding
		} else {
			enc = base32.StdEncoding
		}
		if noPad {
			enc = enc.WithPadding(base32.NoPadding)
		}
		return processBase32Pipe(enc, reader, writer)
	}
	p.UnprocessPipeWithFlags = func(flags *flag.FlagSet, reader io.Reader, writer *io.PipeWriter) (err error) {
		err = unprocessBas32Pipe(base32.HexEncoding, reader, writer)
		if err == nil {
			return
		}
		err = unprocessBas32Pipe(base32.StdEncoding, reader, writer)
		if err == nil {
			return
		}
		err = errors.New("Invalid Base32 data")
		return
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		b32Cmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		b32Cmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Base32 encoding as specified by RFC 4648.\n\n")
			b32Cmd.PrintDefaults()
		}
		b32Cmd.Bool("hex", false, "use \"Extended Hex Alphabet\" defined in RFC 4648")
		b32Cmd.Bool("no-pad", false, "disable padding")
		b32Cmd.Parse(args)
		return b32Cmd
	}
	return
}
