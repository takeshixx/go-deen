package examples

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginExample is a reference plugin demonstrating the unified plugin
// contract for a reversible operation (here: Base64). A plugin provides a
// Process function (forward) and, when the operation is reversible, an
// Unprocess function (reverse). Both read from an io.Reader and write to an
// io.Writer, and must return on the first error.
func NewPluginExample() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "example"
	p.Aliases = []string{".example", "ex", ".ex"}
	p.Category = "codecs"
	p.Description = "Reference plugin that Base64-encodes and decodes its input."
	// Optional: register plugin-specific flags.
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("url", false, "use the URL-safe alphabet")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := base64.StdEncoding
		if helpers.IsBoolFlag(flags, "url") {
			enc = base64.URLEncoding
		}
		encoder := base64.NewEncoder(enc, w)
		if _, err := io.Copy(encoder, r); err != nil {
			return err
		}
		return encoder.Close()
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		enc := base64.StdEncoding
		if helpers.IsBoolFlag(flags, "url") {
			enc = base64.URLEncoding
		}
		_, err := io.Copy(w, base64.NewDecoder(enc, r))
		return err
	}
	return p
}

// NewPluginStreamExample is a reference plugin for a one-way operation (here:
// SHA1). One-way plugins leave Unprocess nil; deen reports an error if a user
// requests the decode direction.
func NewPluginStreamExample() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "streamexample"
	p.Aliases = []string{"strex"}
	p.Category = "hashs"
	p.Description = "Reference one-way plugin that hashes its input with SHA1."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		h := sha1.New()
		if _, err := io.Copy(h, r); err != nil {
			return err
		}
		_, err := io.WriteString(w, hex.EncodeToString(h.Sum(nil)))
		return err
	}
	return p
}
