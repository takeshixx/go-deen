package compressions

import (
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doZlibCompress(task *types.DeenTask, level int) {
	go func() {
		defer task.Close()
		compressor, err := zlib.NewWriterLevel(task.PipeWriter, level)
		if err != nil {
			task.ErrChan <- err
		}
		if _, err := io.Copy(compressor, task.Reader); err != nil {
			task.ErrChan <- err
		}
		err = compressor.Close()
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginZlib creates a new zlib plugin
func NewPluginZlib() (p types.DeenPlugin) {
	p.Name = "zlib"
	p.Aliases = []string{".zlib"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doZlibCompress(task, zlib.DefaultCompression)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		compressionLevel := zlib.DefaultCompression
		level := flags.Lookup("level")
		cliLevel, err := strconv.Atoi(level.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		if cliLevel >= -1 || cliLevel < 10 {
			compressionLevel = cliLevel
		}
		doZlibCompress(task, compressionLevel)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			wrappedReader := types.TrimReader{}
			wrappedReader.Rd = task.Reader
			decompressor, err := zlib.NewReader(wrappedReader)
			if err != nil {
				task.ErrChan <- err
			}
			_, err = io.Copy(task.PipeWriter, decompressor)
			if err != nil {
				task.ErrChan <- err
			}
		}()
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Implements reading and writing of zlib format compressed data (RFC1950).\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			flags.Int("level", zlib.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
		}
		flags.Parse(args)
		return flags
	}
	return
}
