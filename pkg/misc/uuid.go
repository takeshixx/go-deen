package misc

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func parseUUID(s string) ([16]byte, error) {
	var id [16]byte
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "urn:uuid:")
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return id, fmt.Errorf("invalid UUID length")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}
	copy(id[:], b)
	return id, nil
}

func formatUUID(id [16]byte) string {
	h := hex.EncodeToString(id[:])
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

func uuidVariant(id [16]byte) string {
	switch {
	case id[8]&0x80 == 0:
		return "NCS"
	case id[8]&0xc0 == 0x80:
		return "RFC 4122"
	case id[8]&0xe0 == 0xc0:
		return "Microsoft"
	default:
		return "future"
	}
}

func newUUIDv4() ([16]byte, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return id, err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id, nil
}

// NewPluginUUID creates a UUID generation, formatting and inspection plugin.
func NewPluginUUID() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "uuid"
	p.Aliases = []string{"guid"}
	p.Category = "misc"
	p.Description = "Generate UUID v4 values, format raw UUID bytes and inspect UUID text."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("gen", false, "generate a random UUID v4")
		flags.Bool("info", false, "print UUID version and variant")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var id [16]byte
		if helpers.IsBoolFlag(flags, "gen") || len(strings.TrimSpace(string(input))) == 0 {
			id, err = newUUIDv4()
			if err != nil {
				return err
			}
		} else if len(input) == 16 {
			copy(id[:], input)
		} else {
			id, err = parseUUID(string(input))
			if err != nil {
				return err
			}
		}
		if helpers.IsBoolFlag(flags, "info") {
			_, err = fmt.Fprintf(w, "uuid: %s\nversion: %d\nvariant: %s\n", formatUUID(id), id[6]>>4, uuidVariant(id))
			return err
		}
		_, err = io.WriteString(w, formatUUID(id))
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		id, err := parseUUID(string(input))
		if err != nil {
			return err
		}
		_, err = w.Write(id[:])
		return err
	}
	return p
}
