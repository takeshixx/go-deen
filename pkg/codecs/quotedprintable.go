package codecs

import (
	"flag"
	"io"
	"mime/quotedprintable"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginQuotedPrintable creates a quoted-printable plugin (RFC 2045).
func NewPluginQuotedPrintable() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "quoted-printable"
	p.Aliases = []string{".quoted-printable", "qp", ".qp"}
	p.Category = "codecs"
	p.Description = "Quoted-printable encoding as used in MIME email (RFC 2045)."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return encodeStream(r, w, func(w io.Writer) io.WriteCloser { return quotedprintable.NewWriter(w) })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		_, err := io.Copy(w, quotedprintable.NewReader(r))
		return err
	}
	return p
}
