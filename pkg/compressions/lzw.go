package compressions

import (
	"bytes"
	"compress/lzw"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doLzw(reader *io.Reader, order lzw.Order, litWidth int) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	compressor := lzw.NewWriter(&outBuf, order, litWidth)
	if _, err := io.Copy(compressor, *reader); err != nil {
		return outBuf.Bytes(), err
	}
	compressor.Close()
	return outBuf.Bytes(), err
}

func undoLzw(reader *io.Reader, order lzw.Order, litWidth int) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	decompressor := lzw.NewReader(*reader, order, litWidth)
	wrapper := struct{ io.Writer }{&outBuf}
	if _, err := io.Copy(wrapper, decompressor); err != nil {
		return outBuf.Bytes(), err
	}
	return outBuf.Bytes(), err
}

// NewPluginLzw creates a new zlib plugin
func NewPluginLzw() (p types.DeenPlugin) {
	p.Name = "lzw"
	p.Aliases = []string{".lzw"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doLzw(&reader, lzw.LSB, 8)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		orderVal := flags.Lookup("order")
		order, err := strconv.Atoi(orderVal.Value.String())
		if err != nil {
			return []byte{}, err
		}
		widthVal := flags.Lookup("lit-width")
		width, err := strconv.Atoi(widthVal.Value.String())
		if err != nil {
			return []byte{}, err
		}
		return doLzw(&reader, lzw.Order(order), width)
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return undoLzw(&reader, lzw.LSB, 8)
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		orderVal := flags.Lookup("order")
		order, err := strconv.Atoi(orderVal.Value.String())
		if err != nil {
			return []byte{}, err
		}
		widthVal := flags.Lookup("lit-width")
		width, err := strconv.Atoi(widthVal.Value.String())
		if err != nil {
			return []byte{}, err
		}
		return undoLzw(&reader, lzw.Order(order), width)
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		lzwCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		lzwCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Implements the Lempel-Ziv-Welch compressed data format.\n\n")
			lzwCmd.PrintDefaults()
		}
		// TODO: add LSB and MSB values to help page
		lzwCmd.Int("order", int(lzw.LSB), "LSB (GIF) or MSB (TIFF & PDF)")
		lzwCmd.Int("lit-width", 8, "number of output bytes")
		lzwCmd.Parse(args)
		return lzwCmd
	}
	return
}
