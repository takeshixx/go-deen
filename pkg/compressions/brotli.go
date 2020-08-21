package compressions

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"

	"github.com/andybalholm/brotli"
)

func doBrotliCompress(task *types.DeenTask, options brotli.WriterOptions) {
	go func() {
		compressor := brotli.NewWriterOptions(task.PipeWriter, options)
		_, err := io.Copy(compressor, task.Reader)
		if err != nil {
			task.ErrChan <- err
		}
		err = compressor.Close()
		if err != nil {
			task.ErrChan <- err
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

func doBrotliDecompress(task *types.DeenTask) {
	go func() {
		wrappedReader := types.TrimReader{}
		wrappedReader.Rd = task.Reader
		decompressor := brotli.NewReader(wrappedReader)
		_, err := io.Copy(task.PipeWriter, decompressor)
		if err != nil {
			task.ErrChan <- err
		}
		err = task.PipeWriter.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginBrotli creates a new brotli plugin
func NewPluginBrotli() (p types.DeenPlugin) {
	p.Name = "brotli"
	p.Aliases = []string{".brotli", "br", ".br"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		options := &brotli.WriterOptions{
			Quality: brotli.DefaultCompression,
			LGWin:   0,
		}
		doBrotliCompress(task, *options)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		levelFlag := flags.Lookup("level")
		level, err := strconv.Atoi(levelFlag.Value.String())
		if err != nil {
			level = brotli.DefaultCompression
		}
		if level < 0 || level > 11 {
			task.ErrChan <- errors.New("Invalid level")
			return
		}
		lgwinFlag := flags.Lookup("lgwin")
		lgwin, err := strconv.Atoi(lgwinFlag.Value.String())
		if err != nil {
			lgwin = 0
		}
		if lgwin < 0 || lgwin > 24 {
			task.ErrChan <- errors.New("Window size is invalid")
			return
		}
		options := &brotli.WriterOptions{
			Quality: level,
			LGWin:   lgwin,
		}
		doBrotliCompress(task, *options)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		doBrotliDecompress(task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Decription\n\n")
			flags.PrintDefaults()
		}
		flags.Int("level", brotli.DefaultCompression, "compression level")
		flags.Int("lgwin", 0, "sliding window size")
		flags.Parse(args)
		return flags
	}
	return
}
