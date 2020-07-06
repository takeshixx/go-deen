package compressions

import (
	"bytes"
	"compress/flate"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/takeshixx/deen/pkg/types"
)

func doFlate(reader *io.Reader, level int) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	compressor, err := flate.NewWriter(&outBuf, level)
	if err != nil {
		return outBuf.Bytes(), err
	}
	if _, err := io.Copy(compressor, *reader); err != nil {
		return outBuf.Bytes(), err
	}
	compressor.Close()
	return outBuf.Bytes(), err
}

func doDeflate(reader *io.Reader, level int) ([]byte, error) {
	var outBuf bytes.Buffer
	var err error
	decompressor := flate.NewReader(*reader)
	wrapper := struct{ io.Writer }{&outBuf}
	if _, err := io.Copy(wrapper, decompressor); err != nil {
		// TODO: this seems kind of unnecessary...
		return outBuf.Bytes(), err
	}
	return outBuf.Bytes(), err
}

// NewPluginFlate creates a new PluginDeflate object
func NewPluginFlate() (p types.DeenPlugin) {
	p.Name = "flate"
	p.Aliases = []string{".flate"}
	p.Type = "compression"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doFlate(&reader, flate.DefaultCompression)
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressionLevel := flate.DefaultCompression
		level := flags.Lookup("level")
		cliLevel, err := strconv.Atoi(level.Value.String())
		if err != nil {
			return outBuf.Bytes(), err
		}
		if cliLevel >= -1 || cliLevel < 10 {
			compressionLevel = cliLevel
		}
		return doFlate(&reader, compressionLevel)
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		return doDeflate(&reader, flate.DefaultCompression)
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error
		compressionLevel := flate.DefaultCompression
		level := flags.Lookup("level")
		cliLevel, err := strconv.Atoi(level.Value.String())
		if err != nil {
			return outBuf.Bytes(), err
		}
		if cliLevel >= -1 || cliLevel < 10 {
			compressionLevel = cliLevel
		}
		return doDeflate(&reader, compressionLevel)
	}
	p.AddCliOptionsFunc = func(self *types.DeenPlugin, args []string) *flag.FlagSet {
		deflateCmd := flag.NewFlagSet(p.Name, flag.ExitOnError)
		deflateCmd.Usage = func() {
			fmt.Printf("Usage of %s:\n\n", p.Name)
			fmt.Printf("Implements the DEFLATE compressed data format (RFC1951).\n\n")
			deflateCmd.PrintDefaults()
		}
		levelDescription := "compression level\n" +
			"  No compression:\t" + strconv.Itoa(flate.NoCompression) + "\n" +
			"  Best speed:\t\t" + strconv.Itoa(flate.BestSpeed) + "\n" +
			"  Best compression:\t" + strconv.Itoa(flate.BestCompression) + "\n" +
			"  Default compression:\t" + strconv.Itoa(flate.DefaultCompression) + "\n "
		deflateCmd.Int("level", flate.DefaultCompression, levelDescription)
		deflateCmd.Parse(args)
		return deflateCmd
	}
	return
}
