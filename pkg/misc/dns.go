package misc

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

func encodeDNSName(name string) ([]byte, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, ".")
	if name == "" {
		return []byte{0}, nil
	}
	var out []byte
	for _, label := range strings.Split(name, ".") {
		if label == "" {
			return nil, fmt.Errorf("empty DNS label")
		}
		if len(label) > 63 {
			return nil, fmt.Errorf("DNS label too long")
		}
		out = append(out, byte(len(label)))
		out = append(out, label...)
	}
	if len(out)+1 > 255 {
		return nil, fmt.Errorf("DNS name too long")
	}
	out = append(out, 0)
	return out, nil
}

func decodeDNSName(data []byte) (string, error) {
	var labels []string
	for pos := 0; ; {
		if pos >= len(data) {
			return "", fmt.Errorf("truncated DNS name")
		}
		l := int(data[pos])
		pos++
		if l == 0 {
			break
		}
		if l&0xc0 != 0 {
			return "", fmt.Errorf("compressed DNS pointers are not supported in name-only mode")
		}
		if l > 63 || pos+l > len(data) {
			return "", fmt.Errorf("invalid DNS label length")
		}
		labels = append(labels, string(data[pos:pos+l]))
		pos += l
	}
	return strings.Join(labels, ".") + ".", nil
}

// NewPluginDNS creates a DNS name wire-format codec.
func NewPluginDNS() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "dns"
	p.Aliases = []string{"dnsname"}
	p.Category = "misc"
	p.Description = "Encode and decode DNS names in wire format."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		out, err := encodeDNSName(string(input))
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		name, err := decodeDNSName(bytes.TrimSpace(input))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, name)
		return err
	}
	return p
}
