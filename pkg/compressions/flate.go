package compressions

import (
	"compress/flate"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doFlate(task *types.DeenTask, level int) {
	go func() {
		defer task.Close()
		compressor, err := flate.NewWriter(task.PipeWriter, level)
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

func doDeflate(task *types.DeenTask) {
	go func() {
		defer task.Close()
		wrappedReader := types.TrimReader{}
		wrappedReader.Rd = task.Reader
		decompressor := flate.NewReader(wrappedReader)
		_, err := io.Copy(task.PipeWriter, decompressor)
		if err != nil {
			task.ErrChan <- err
		}
	}()
}

// NewPluginFlate creates a new PluginDeflate object
func NewPluginFlate() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "flate"
	p.Aliases = []string{".flate"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doFlate(task, flate.DefaultCompression)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		compressionLevel := flate.DefaultCompression
		level := flags.Lookup("level")
		cliLevel, err := strconv.Atoi(level.Value.String())
		if err != nil {
			task.ErrChan <- err
		}
		if cliLevel >= -1 || cliLevel < 10 {
			compressionLevel = cliLevel
		}
		doFlate(task, compressionLevel)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		doDeflate(task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Implements the DEFLATE compressed data format (RFC1951).\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			levelDescription := "compression level\n" +
				"  No compression:\t" + strconv.Itoa(flate.NoCompression) + "\n" +
				"  Best speed:\t\t" + strconv.Itoa(flate.BestSpeed) + "\n" +
				"  Best compression:\t" + strconv.Itoa(flate.BestCompression) + "\n" +
				"  Default compression:\t" + strconv.Itoa(flate.DefaultCompression) + "\n "
			flags.Int("level", flate.DefaultCompression, levelDescription)
		}
		flags.Parse(args)
		return flags
	}
	return
}
