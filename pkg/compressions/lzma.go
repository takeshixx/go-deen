package compressions

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ulikunitz/xz/lzma"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginLZMA creates a new PluginLZMA object
func NewPluginLZMA() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "lzma"
	p.Aliases = []string{".lzma"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			compressor, err := lzma.NewWriter(task.PipeWriter)
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
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.ProcessDeenTaskFunc(task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			decompressor, err := lzma.NewReader(task.Reader)
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
			fmt.Fprintf(os.Stderr, "Decoding and encoding of LZMA streams.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginLZMA2 creates a new PluginLZMA2 object
func NewPluginLZMA2() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "lzma2"
	p.Aliases = []string{".lzma2"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			compressor, err := lzma.NewWriter2(task.PipeWriter)
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
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.ProcessDeenTaskFunc(task)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			decompressor, err := lzma.NewReader2(task.Reader)
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
			fmt.Fprintf(os.Stderr, "Decoding and encoding of LZMA2 streams.\n\n")
			flags.PrintDefaults()
		}
		flags.Parse(args)
		return flags
	}
	return
}
