package compressions

import (
	"compress/lzw"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doLzwCompress(task *types.DeenTask, order lzw.Order, litWidth int) {
	go func() {
		defer task.Close()
		compressor := lzw.NewWriter(task.PipeWriter, order, litWidth)
		if _, err := io.Copy(compressor, task.Reader); err != nil {
			task.ErrChan <- err
		}
		err := compressor.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

func doLzwUncompress(task *types.DeenTask, order lzw.Order, litWidth int) {
	go func() {
		defer task.Close()
		wrappedReader := types.TrimReader{}
		wrappedReader.Rd = task.Reader
		decompressor := lzw.NewReader(wrappedReader, order, litWidth)
		_, err := io.Copy(task.PipeWriter, decompressor)
		if err != nil {
			task.ErrChan <- err
		}
		err = decompressor.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginLzw creates a new zlib plugin
func NewPluginLzw() (p types.DeenPlugin) {
	p.Name = "lzw"
	p.Aliases = []string{".lzw"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doLzwCompress(task, lzw.LSB, 8)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		orderVal := flags.Lookup("order")
		order, err := strconv.Atoi(orderVal.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		widthVal := flags.Lookup("lit-width")
		width, err := strconv.Atoi(widthVal.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		doLzwCompress(task, lzw.Order(order), width)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		doLzwUncompress(task, lzw.LSB, 8)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		orderVal := flags.Lookup("order")
		order, err := strconv.Atoi(orderVal.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		widthVal := flags.Lookup("lit-width")
		width, err := strconv.Atoi(widthVal.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		doLzwUncompress(task, lzw.Order(order), width)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Implements the Lempel-Ziv-Welch compressed data format.\n\n")
			flags.PrintDefaults()
		}
		// TODO: add LSB and MSB values to help page
		flags.Int("order", int(lzw.LSB), "LSB (GIF) or MSB (TIFF & PDF)")
		flags.Int("lit-width", 8, "number of output bytes")
		flags.Parse(args)
		return flags
	}
	return
}
