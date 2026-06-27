package formatters

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/takeshixx/deen/pkg/types"
)

const maxProtoDepth = 4

func NewPluginProtobuf() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "protobuf"
	p.Aliases = []string{"proto", "pb"}
	p.Category = "formatters"
	p.Description = "Decodes schema-less Protocol Buffers wire data into a readable field listing."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var out strings.Builder
		if err := decodeProtoMessage(data, &out, 0); err != nil {
			return err
		}
		_, err = io.WriteString(w, out.String())
		return err
	}
	return p
}

func decodeProtoMessage(data []byte, out *strings.Builder, depth int) error {
	for len(data) > 0 {
		key, n := binary.Uvarint(data)
		if n <= 0 {
			return fmt.Errorf("invalid protobuf key")
		}
		data = data[n:]

		field := key >> 3
		wire := key & 0x7
		if field == 0 {
			return fmt.Errorf("invalid field number 0")
		}

		indent(out, depth)
		fmt.Fprintf(out, "%d: ", field)

		switch wire {
		case 0:
			v, n := binary.Uvarint(data)
			if n <= 0 {
				return fmt.Errorf("field %d: invalid varint", field)
			}
			data = data[n:]
			fmt.Fprintf(out, "varint %d", v)
			if sv := decodeZigZag(v); sv != int64(v) {
				fmt.Fprintf(out, " (zigzag %d)", sv)
			}
			out.WriteByte('\n')
		case 1:
			if len(data) < 8 {
				return fmt.Errorf("field %d: truncated fixed64", field)
			}
			v := binary.LittleEndian.Uint64(data[:8])
			data = data[8:]
			fmt.Fprintf(out, "fixed64 0x%016x (%d)\n", v, v)
		case 2:
			l, n := binary.Uvarint(data)
			if n <= 0 {
				return fmt.Errorf("field %d: invalid length", field)
			}
			data = data[n:]
			if uint64(len(data)) < l {
				return fmt.Errorf("field %d: length %d exceeds remaining input", field, l)
			}
			value := data[:l]
			data = data[l:]
			writeLengthDelimited(out, depth, value)
		case 3:
			out.WriteString("start-group\n")
		case 4:
			out.WriteString("end-group\n")
		case 5:
			if len(data) < 4 {
				return fmt.Errorf("field %d: truncated fixed32", field)
			}
			v := binary.LittleEndian.Uint32(data[:4])
			data = data[4:]
			fmt.Fprintf(out, "fixed32 0x%08x (%d)\n", v, v)
		default:
			return fmt.Errorf("field %d: unknown wire type %d", field, wire)
		}
	}
	return nil
}

func writeLengthDelimited(out *strings.Builder, depth int, value []byte) {
	if len(value) == 0 {
		out.WriteString("length 0 bytes\n")
		return
	}
	if utf8.Valid(value) && isMostlyPrintable(value) {
		fmt.Fprintf(out, "string %s\n", strconv.Quote(string(value)))
		return
	}
	if depth < maxProtoDepth && looksLikeProtoMessage(value) {
		fmt.Fprintf(out, "message %d bytes {\n", len(value))
		if err := decodeProtoMessage(value, out, depth+1); err != nil {
			indent(out, depth+1)
			fmt.Fprintf(out, "bytes %s\n", hex.EncodeToString(value))
		}
		indent(out, depth)
		out.WriteString("}\n")
		return
	}
	fmt.Fprintf(out, "bytes %s\n", hex.EncodeToString(value))
}

func looksLikeProtoMessage(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	var out strings.Builder
	return decodeProtoMessage(data, &out, maxProtoDepth) == nil
}

func isMostlyPrintable(data []byte) bool {
	var printable, total int
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			return false
		}
		total++
		if r == '\n' || r == '\r' || r == '\t' || (r >= 0x20 && r != 0x7f) {
			printable++
		}
		data = data[size:]
	}
	return total > 0 && printable*100/total >= 85
}

func decodeZigZag(v uint64) int64 {
	return int64((v >> 1) ^ uint64(-int64(v&1)))
}

func indent(out *strings.Builder, depth int) {
	out.Write(bytes.Repeat([]byte("  "), depth))
}
