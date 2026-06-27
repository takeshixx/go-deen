package misc

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

type asn1Node struct {
	Offset      int
	HeaderLen   int
	Length      int
	Class       int
	Constructed bool
	Tag         int
	Content     []byte
	Children    []asn1Node
}

var asn1UniversalNames = map[int]string{
	1:  "BOOLEAN",
	2:  "INTEGER",
	3:  "BIT STRING",
	4:  "OCTET STRING",
	5:  "NULL",
	6:  "OBJECT IDENTIFIER",
	12: "UTF8String",
	16: "SEQUENCE",
	17: "SET",
	19: "PrintableString",
	20: "T61String",
	22: "IA5String",
	23: "UTCTime",
	24: "GeneralizedTime",
	30: "BMPString",
}

func parseASN1DER(data []byte, base int) ([]asn1Node, error) {
	var nodes []asn1Node
	for pos := 0; pos < len(data); {
		node, consumed, err := parseASN1Node(data[pos:], base+pos)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		pos += consumed
	}
	return nodes, nil
}

func parseASN1Node(data []byte, offset int) (asn1Node, int, error) {
	if len(data) < 2 {
		return asn1Node{}, 0, fmt.Errorf("truncated ASN.1 element at offset %d", offset)
	}
	pos := 0
	first := data[pos]
	pos++
	node := asn1Node{
		Offset:      offset,
		Class:       int(first >> 6),
		Constructed: first&0x20 != 0,
		Tag:         int(first & 0x1f),
	}
	if node.Tag == 0x1f {
		tag := 0
		for {
			if pos >= len(data) {
				return asn1Node{}, 0, fmt.Errorf("truncated high-tag-number at offset %d", offset)
			}
			b := data[pos]
			pos++
			tag = tag<<7 | int(b&0x7f)
			if b&0x80 == 0 {
				break
			}
		}
		node.Tag = tag
	}
	if pos >= len(data) {
		return asn1Node{}, 0, fmt.Errorf("missing length at offset %d", offset)
	}
	lengthByte := data[pos]
	pos++
	switch {
	case lengthByte&0x80 == 0:
		node.Length = int(lengthByte)
	case lengthByte == 0x80:
		return asn1Node{}, 0, fmt.Errorf("indefinite lengths are BER, not DER, at offset %d", offset)
	default:
		n := int(lengthByte & 0x7f)
		if n == 0 || n > 4 {
			return asn1Node{}, 0, fmt.Errorf("invalid length at offset %d", offset)
		}
		if pos+n > len(data) {
			return asn1Node{}, 0, fmt.Errorf("truncated length at offset %d", offset)
		}
		for i := 0; i < n; i++ {
			node.Length = node.Length<<8 | int(data[pos+i])
		}
		pos += n
	}
	if pos+node.Length > len(data) {
		return asn1Node{}, 0, fmt.Errorf("content overruns input at offset %d", offset)
	}
	node.HeaderLen = pos
	node.Content = data[pos : pos+node.Length]
	if node.Constructed {
		children, err := parseASN1DER(node.Content, offset+node.HeaderLen)
		if err != nil {
			return asn1Node{}, 0, err
		}
		node.Children = children
	}
	return node, node.HeaderLen + node.Length, nil
}

func renderASN1(nodes []asn1Node, w io.Writer, level int) {
	for _, node := range nodes {
		indent := strings.Repeat("    ", level)
		fmt.Fprintf(w, "%s@%04x %s tag=%d len=%d header=%d", indent, node.Offset, asn1ClassName(node.Class), node.Tag, node.Length, node.HeaderLen)
		if node.Constructed {
			fmt.Fprint(w, " constructed")
		}
		if node.Class == 0 {
			if name := asn1UniversalNames[node.Tag]; name != "" {
				fmt.Fprintf(w, " %s", name)
			}
		}
		if value := asn1PrimitiveValue(node); value != "" {
			fmt.Fprintf(w, " = %s", value)
		}
		fmt.Fprintln(w)
		if len(node.Children) > 0 {
			renderASN1(node.Children, w, level+1)
		}
	}
}

func asn1ClassName(class int) string {
	switch class {
	case 0:
		return "universal"
	case 1:
		return "application"
	case 2:
		return "context"
	case 3:
		return "private"
	default:
		return "unknown"
	}
}

func asn1PrimitiveValue(node asn1Node) string {
	if node.Constructed || node.Class != 0 {
		return ""
	}
	switch node.Tag {
	case 1:
		if len(node.Content) != 1 {
			return "invalid BOOLEAN"
		}
		return strconv.FormatBool(node.Content[0] != 0)
	case 2:
		return signedASN1Integer(node.Content).String()
	case 3:
		if len(node.Content) == 0 {
			return "invalid BIT STRING"
		}
		return fmt.Sprintf("unused-bits=%d bytes=%s", node.Content[0], compactHex(node.Content[1:]))
	case 4:
		return "bytes=" + compactHex(node.Content)
	case 5:
		if len(node.Content) == 0 {
			return "NULL"
		}
		return "invalid NULL"
	case 6:
		if oid, ok := parseASN1OID(node.Content); ok {
			return oid
		}
		return "invalid OBJECT IDENTIFIER"
	case 12, 19, 20, 22, 23, 24:
		return strconv.Quote(string(node.Content))
	default:
		return ""
	}
}

func signedASN1Integer(data []byte) *big.Int {
	if len(data) == 0 {
		return big.NewInt(0)
	}
	n := new(big.Int).SetBytes(data)
	if data[0]&0x80 == 0 {
		return n
	}
	limit := new(big.Int).Lsh(big.NewInt(1), uint(len(data)*8))
	return n.Sub(n, limit)
}

func parseASN1OID(data []byte) (string, bool) {
	if len(data) == 0 {
		return "", false
	}
	first := int(data[0])
	parts := []int{first / 40, first % 40}
	if parts[0] > 2 {
		parts[0] = 2
		parts[1] = first - 80
	}
	value := 0
	for _, b := range data[1:] {
		value = value<<7 | int(b&0x7f)
		if b&0x80 == 0 {
			parts = append(parts, value)
			value = 0
		}
	}
	if value != 0 {
		return "", false
	}
	var out strings.Builder
	for i, part := range parts {
		if i > 0 {
			out.WriteByte('.')
		}
		out.WriteString(strconv.Itoa(part))
	}
	return out.String(), true
}

func compactHex(data []byte) string {
	if len(data) == 0 {
		return "<empty>"
	}
	const max = 24
	if len(data) <= max {
		return hex.EncodeToString(data)
	}
	return hex.EncodeToString(data[:max]) + fmt.Sprintf("... (%d bytes)", len(data))
}

func derFromInput(input []byte) []byte {
	var blocks [][]byte
	for {
		block, rest := pem.Decode(input)
		if block == nil {
			break
		}
		blocks = append(blocks, block.Bytes)
		input = rest
	}
	if len(blocks) == 0 {
		return input
	}
	return bytes.Join(blocks, nil)
}

// NewPluginASN1 creates a schema-less ASN.1 DER tree printer.
func NewPluginASN1() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "asn1"
	p.Aliases = []string{"der"}
	p.Category = "misc"
	p.Description = "Parse ASN.1 DER or PEM input into a readable tag/length/value tree."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		der := derFromInput(input)
		if len(der) == 0 {
			return fmt.Errorf("empty DER input")
		}
		nodes, err := parseASN1DER(der, 0)
		if err != nil {
			return err
		}
		renderASN1(nodes, w, 0)
		return nil
	}
	return p
}
