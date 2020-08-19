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
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		processBase32(base32.StdEncoding, task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		unprocessBas32(base32.StdEncoding, task)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
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
		processBase32(enc, task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		var enc *base32.Encoding
		if hex, err := isHex(flags); hex && err == nil {
			enc = base32.HexEncoding
		} else {
			enc = base32.StdEncoding
		}
		unprocessBas32(enc, task)
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
