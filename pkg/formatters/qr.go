package formatters

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"

	"github.com/liyue201/goqr"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginQR creates a QR PNG encoder and image decoder.
func NewPluginQR() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "qr"
	p.Aliases = []string{"qrcode"}
	p.Category = "formatters"
	p.Description = "Encode text as QR PNG and decode QR images back to text."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Int("size", 256, "QR image size in pixels")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		size := helpers.IntFlag(flags, "size", 256)
		if size <= 0 {
			return fmt.Errorf("size must be positive")
		}
		png, err := qrcode.Encode(strings.TrimRight(string(input), "\r\n"), qrcode.Medium, size)
		if err != nil {
			return err
		}
		_, err = w.Write(png)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		img, _, err := image.Decode(r)
		if err != nil {
			return err
		}
		codes, err := goqr.Recognize(img)
		if err != nil {
			return err
		}
		if len(codes) == 0 {
			return fmt.Errorf("no QR code found")
		}
		for i, code := range codes {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprint(w, string(code.Payload))
		}
		return nil
	}
	return p
}
