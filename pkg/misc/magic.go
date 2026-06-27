package misc

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/takeshixx/deen/pkg/types"
)

type magicSignature struct {
	name string
	mime string
	pat  []byte
}

var magicSignatures = []magicSignature{
	{"PNG image", "image/png", []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}},
	{"JPEG image", "image/jpeg", []byte{0xff, 0xd8, 0xff}},
	{"GIF image", "image/gif", []byte("GIF8")},
	{"PDF document", "application/pdf", []byte("%PDF-")},
	{"ZIP archive", "application/zip", []byte("PK\x03\x04")},
	{"gzip data", "application/gzip", []byte{0x1f, 0x8b}},
	{"bzip2 data", "application/x-bzip2", []byte("BZh")},
	{"zstd data", "application/zstd", []byte{0x28, 0xb5, 0x2f, 0xfd}},
	{"ELF executable", "application/x-elf", []byte{0x7f, 'E', 'L', 'F'}},
	{"PE executable", "application/vnd.microsoft.portable-executable", []byte("MZ")},
}

// NewPluginMagic creates a simple file signature detector.
func NewPluginMagic() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "magic"
	p.Aliases = []string{"filetype"}
	p.Category = "misc"
	p.Description = "Detect common file types from magic bytes and content sniffing."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		for _, sig := range magicSignatures {
			if bytes.HasPrefix(data, sig.pat) {
				_, err = fmt.Fprintf(w, "type: %s\nmime: %s\n", sig.name, sig.mime)
				return err
			}
		}
		mime := http.DetectContentType(data)
		_, err = fmt.Fprintf(w, "type: unknown\nmime: %s\n", mime)
		return err
	}
	return p
}
