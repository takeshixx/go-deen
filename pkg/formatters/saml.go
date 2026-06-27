package formatters

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSAML creates a formatter for common SAML request/response payloads.
func NewPluginSAML() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "saml"
	p.Aliases = []string{".saml"}
	p.Category = "formatters"
	p.Description = "Decode SAMLRequest/SAMLResponse payloads to XML and encode XML back to SAML payloads."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("deflate", false, "DEFLATE-compress before encoding; require DEFLATE when decoding")
		flags.Bool("plain", false, "do not try DEFLATE decompression when decoding")
		flags.Bool("url", false, "URL-escape encoded output; URL-unescape input before decoding")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		payload, err := decodeSAMLPayload(input, helpers.IsBoolFlag(flags, "deflate"), helpers.IsBoolFlag(flags, "plain"), helpers.IsBoolFlag(flags, "url"))
		if err != nil {
			return err
		}
		return reformatXML(bytes.NewReader(payload), w, "    ")
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		var xmlBuf bytes.Buffer
		if err := reformatXML(r, &xmlBuf, ""); err != nil {
			return err
		}
		payload := xmlBuf.Bytes()
		if helpers.IsBoolFlag(flags, "deflate") {
			payload = deflateRaw(payload)
		}
		out := base64.StdEncoding.EncodeToString(payload)
		if helpers.IsBoolFlag(flags, "url") {
			out = url.QueryEscape(out)
		}
		_, err := io.WriteString(w, out)
		return err
	}
	return p
}

func decodeSAMLPayload(input []byte, requireDeflate, plain, shouldUnescape bool) ([]byte, error) {
	s, extracted := extractSAMLValue(strings.TrimSpace(string(input)))
	if !extracted && (shouldUnescape || strings.Contains(s, "%")) {
		if unescaped, err := url.QueryUnescape(s); err == nil {
			s = unescaped
		}
	}
	decoded, err := decodeSAMLBase64(s)
	if err != nil {
		return nil, err
	}
	if plain {
		return decoded, nil
	}
	inflated, err := inflateRaw(decoded)
	if err == nil {
		return inflated, nil
	}
	if requireDeflate {
		return nil, fmt.Errorf("could not DEFLATE-decompress SAML payload: %w", err)
	}
	return decoded, nil
}

func extractSAMLValue(s string) (string, bool) {
	if strings.Contains(s, "SAMLRequest=") || strings.Contains(s, "SAMLResponse=") {
		if values, err := url.ParseQuery(s); err == nil {
			for _, key := range []string{"SAMLRequest", "SAMLResponse"} {
				if v := values.Get(key); v != "" {
					return v, true
				}
			}
		}
	}
	return s, false
}

func decodeSAMLBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	candidates := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range candidates {
		if decoded, err := enc.DecodeString(s); err == nil {
			return decoded, nil
		} else {
			lastErr = err
		}
	}
	return nil, fmt.Errorf("could not decode SAML base64 payload: %w", lastErr)
}

func inflateRaw(data []byte) ([]byte, error) {
	rc := flate.NewReader(bytes.NewReader(data))
	defer rc.Close()
	return io.ReadAll(rc)
}

func deflateRaw(data []byte) []byte {
	var out bytes.Buffer
	w, err := flate.NewWriter(&out, flate.DefaultCompression)
	if err != nil {
		return data
	}
	_, _ = w.Write(data)
	_ = w.Close()
	return out.Bytes()
}
