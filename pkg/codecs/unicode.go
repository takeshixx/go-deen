package codecs

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

func doTransformer(e encoding.Encoding, t *types.DeenTask) {
	go func() {
		defer t.Close()
		encoder := e.NewEncoder()
		writer := encoder.Writer(t.PipeWriter)
		_, err := io.Copy(writer, t.Reader)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Copying into encoder failed")
		}
	}()
}

func undoTransformer(e encoding.Encoding, t *types.DeenTask) {
	go func() {
		defer t.Close()
		decoder := e.NewDecoder()
		reader := decoder.Reader(t.Reader)
		_, err := io.Copy(t.PipeWriter, reader)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Copying into decoder failed")
		}
	}()
}

func readEndianess(flags *flag.FlagSet) (ret unicode.Endianness) {
	ret = unicode.LittleEndian
	if endianFlag := flags.Lookup("big"); endianFlag != nil {
		if bigEndian, err := strconv.ParseBool(endianFlag.Value.String()); err == nil && bigEndian {
			ret = unicode.BigEndian
		}
	}
	return
}

func readBOMPolicy(flags *flag.FlagSet) (ret unicode.BOMPolicy) {
	ret = unicode.IgnoreBOM
	if bomFlag := flags.Lookup("bom"); bomFlag != nil {
		bomMode := bomFlag.Value.String()
		if bomMode == "use" {
			ret = unicode.UseBOM
		} else if bomMode == "ignore" {
			ret = unicode.IgnoreBOM
		} else if bomMode == "expect" {
			ret = unicode.ExpectBOM
		}
	}
	return
}

// NewPluginUnicode creates a new PluginUnicode object
func NewPluginUnicode() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "unicode"
	p.Aliases = []string{".unicode", "utf8", ".utf8", "utf16", ".utf16", "utf32", ".utf32", "euckr", ".euckr"}
	p.Type = "codec"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		doTransformer(unicode.UTF8, task)
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		endianess := readEndianess(flags)
		bomPolicy := readBOMPolicy(flags)
		switch task.Command {
		case "utf8":
			doTransformer(unicode.UTF8, task)
		case "utf16":
			doTransformer(unicode.UTF16(endianess, bomPolicy), task)
		case "utf32":
			doTransformer(utf32.UTF32(utf32.Endianness(endianess), utf32.BOMPolicy(bomPolicy)), task)
		case "euckr":
			doTransformer(korean.EUCKR, task)
		default:
			doTransformer(unicode.UTF8, task)
		}
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		undoTransformer(unicode.UTF8, task)
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		endianess := readEndianess(flags)
		bomPolicy := readBOMPolicy(flags)
		switch task.Command {
		case "utf8":
			undoTransformer(unicode.UTF8, task)
		case "utf16":
			undoTransformer(unicode.UTF16(endianess, bomPolicy), task)
		case "utf32":
			undoTransformer(utf32.UTF32(utf32.Endianness(endianess), utf32.BOMPolicy(bomPolicy)), task)
		case "euckr":
			undoTransformer(korean.EUCKR, task)
		default:
			undoTransformer(unicode.UTF8, task)
		}
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)

		if self.Command == "utf8" {
			flags.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
				fmt.Fprintf(os.Stderr, "Encode/Decode unicode to UTF8.\n\n")
				flags.PrintDefaults()
			}
			flags.Bool("bom", false, "add/strip leading byte order mark")
		} else if self.Command == "utf16" {
			flags.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
				fmt.Fprintf(os.Stderr, "Encode/Decode unicode to UTF16.\n\n")
				flags.PrintDefaults()
			}
			flags.String("bom", "ignore", "BOM mode (use, ignore, expect)")
			flags.Bool("big", false, "use big endian (default: little)")
		} else if self.Command == "utf32" {
			flags.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
				fmt.Fprintf(os.Stderr, "Encode/Decode unicode to UTF32.\n\n")
				flags.PrintDefaults()
			}
			flags.String("bom", "ignore", "BOM mode (use, ignore, expect)")
			flags.Bool("big", false, "use big endian (default: little)")
		} else if self.Command == "euckr" {
			flags.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
				fmt.Fprintf(os.Stderr, "Encode/Decode unicode to EUC-KR encoding, also known as Code Page 949.\n\n")
				flags.PrintDefaults()
			}
		}
		flags.Parse(args)
		return flags
	}
	return
}
