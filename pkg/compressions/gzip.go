package compressions

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doGzipCompress(task *types.DeenTask, level int) {
	defer task.Close()
	compressor, err := gzip.NewWriterLevel(task.PipeWriter, level)
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
}

// NewPluginGzip creates a new zlib plugin
func NewPluginGzip() (p types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "gzip"
	p.Aliases = []string{".gzip"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doGzipCompress(task, gzip.DefaultCompression)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		compressionLevel := gzip.DefaultCompression
		level := flags.Lookup("level")
		cliLevel, err := strconv.Atoi(level.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		if cliLevel >= -1 || cliLevel < 10 {
			compressionLevel = cliLevel
		}
		doGzipCompress(task, compressionLevel)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		defer task.Close()
		wrappedReader := types.TrimReader{}
		wrappedReader.Rd = task.Reader
		decompressor, err := gzip.NewReader(wrappedReader)
		if err != nil {
			task.ErrChan <- err
		}
		_, err = io.Copy(task.PipeWriter, decompressor)
		if err != nil {
			task.ErrChan <- err
		}
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Implements reading and writing of gzip format compressed files (RFC1952).\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			flags.Int("level", gzip.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
		}
		flags.Parse(args)
		return flags
	}
	return
}
