package codecs

import (
	"flag"
	"io"
	"mime/quotedprintable"

	"github.com/takeshixx/deen/pkg/types"
)

const qpHex = "0123456789ABCDEF"

// encodeQuotedPrintable writes a binary-safe quoted-printable encoding of data.
// Unlike the stdlib quotedprintable.Writer (which normalises CR/LF and is meant
// for text), this escapes every byte that is not a literal printable ASCII
// character, so arbitrary binary round-trips through the stdlib decoder. Output
// is wrapped with soft line breaks to respect the 76-character line limit.
func encodeQuotedPrintable(w io.Writer, data []byte) error {
	lineLen := 0
	// emit writes a token (1 or 3 bytes) inserting a soft line break first if
	// the token would not fit within the 76-character line limit (reserving one
	// column for the trailing "=" of a soft break).
	emit := func(token []byte) error {
		if lineLen+len(token) > 75 {
			if _, err := io.WriteString(w, "=\r\n"); err != nil {
				return err
			}
			lineLen = 0
		}
		if _, err := w.Write(token); err != nil {
			return err
		}
		lineLen += len(token)
		return nil
	}
	for _, b := range data {
		// Printable ASCII (33-126) except '=' is written literally.
		if b >= 33 && b <= 126 && b != '=' {
			if err := emit([]byte{b}); err != nil {
				return err
			}
			continue
		}
		if err := emit([]byte{'=', qpHex[b>>4], qpHex[b&0x0f]}); err != nil {
			return err
		}
	}
	return nil
}

// NewPluginQuotedPrintable creates a quoted-printable plugin (RFC 2045).
func NewPluginQuotedPrintable() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "quoted-printable"
	p.Aliases = []string{".quoted-printable", "qp", ".qp"}
	p.Category = "codecs"
	p.Description = "Quoted-printable encoding as used in MIME email (RFC 2045).\nEncoding is binary-safe; decoding accepts any valid quoted-printable."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return encodeQuotedPrintable(w, data)
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		_, err := io.Copy(w, quotedprintable.NewReader(r))
		return err
	}
	return p
}
