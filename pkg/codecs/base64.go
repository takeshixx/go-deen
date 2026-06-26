package codecs

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// base64Encoding selects the encoding for the forward direction based on flags.
func base64Encoding(flags *flag.FlagSet) *base64.Encoding {
	url := helpers.IsBoolFlag(flags, "url")
	raw := helpers.IsBoolFlag(flags, "raw")
	switch {
	case url && raw:
		return base64.RawURLEncoding
	case url:
		return base64.URLEncoding
	case raw:
		return base64.RawStdEncoding
	default:
		return base64.StdEncoding
	}
}

// NewPluginBase64 creates a new base64 plugin (RFC 4648).
func NewPluginBase64() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "base64"
	p.Aliases = []string{".base64", "b64", ".b64"}
	p.Category = "codecs"
	p.Description = "Base64 encoding as defined in RFC 4648. By default decoding\nattempts the standard, raw, URL and raw-URL alphabets in turn."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("strict", false, "only use standard Base64 (no alternate alphabets when decoding)")
		flags.Bool("raw", false, "unpadded Base64 encoding (RFC 4648 section 3.2)")
		flags.Bool("url", false, "URL-safe Base64 alphabet")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := base64Encoding(flags)
		return encodeStream(r, w, func(w io.Writer) io.WriteCloser { return base64.NewEncoder(enc, w) })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		data = bytes.TrimSpace(data)

		var candidates []*base64.Encoding
		switch {
		case helpers.IsBoolFlag(flags, "strict"):
			candidates = []*base64.Encoding{base64.StdEncoding}
		case helpers.IsBoolFlag(flags, "url") || helpers.IsBoolFlag(flags, "raw"):
			candidates = []*base64.Encoding{base64Encoding(flags)}
		default:
			candidates = []*base64.Encoding{
				base64.StdEncoding, base64.RawStdEncoding,
				base64.URLEncoding, base64.RawURLEncoding,
			}
		}

		var lastErr error
		for _, enc := range candidates {
			if decoded, derr := enc.DecodeString(string(data)); derr == nil {
				_, err = w.Write(decoded)
				return err
			} else {
				lastErr = derr
			}
		}
		return fmt.Errorf("could not decode Base64 input: %w", lastErr)
	}
	return p
}
