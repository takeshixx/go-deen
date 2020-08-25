package compressions

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/dsnet/compress/bzip2"
	"github.com/takeshixx/deen/pkg/types"
)

func doBZip2Compress(task *types.DeenTask, level int) {
	defer task.Close()
	config := &bzip2.WriterConfig{Level: level}
	compressor, err := bzip2.NewWriter(task.PipeWriter, config)
	if err != nil {
		task.ErrChan <- err
	}
	_, err = io.Copy(compressor, task.Reader)
	if err != nil {
		task.ErrChan <- err
	}
	err = compressor.Close()
	if err != nil {
		task.ErrChan <- err
	}
}

// NewPluginBzip2 creates a new zlib plugin
func NewPluginBzip2() (p types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "bzip2"
	p.Aliases = []string{".bzip2"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doBZip2Compress(task, bzip2.DefaultCompression)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		levelFlag := flags.Lookup("level")
		level, err := strconv.Atoi(levelFlag.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		if level < bzip2.BestSpeed || level > bzip2.BestCompression {
			task.ErrChan <- errors.New("Invalid level")
		}
		doBZip2Compress(task, level)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		defer task.Close()
		wrappedReader := types.TrimReader{}
		wrappedReader.Rd = task.Reader
		decompressor, err := bzip2.NewReader(wrappedReader, nil)
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
			fmt.Fprintf(os.Stderr, "BZip2 compressed data format.\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			flags.Int("level", bzip2.DefaultCompression, "compression level from 1 (best speed) to 9 (best compression)")
		}
		flags.Parse(args)
		return flags
	}
	return
}
