package codecs

import (
	"flag"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// rot13 rotates ASCII letters by 13 positions. It is its own inverse.
func rot13(data []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		switch {
		case b >= 'A' && b <= 'Z':
			out[i] = 'A' + (b-'A'+13)%26
		case b >= 'a' && b <= 'z':
			out[i] = 'a' + (b-'a'+13)%26
		default:
			out[i] = b
		}
	}
	return out
}

// NewPluginROT13 creates a ROT13 plugin. ROT13 is symmetric, so processing and
// unprocessing apply the same transformation.
func NewPluginROT13() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "rot13"
	p.Aliases = []string{".rot13"}
	p.Category = "codecs"
	p.Description = "ROT13 letter substitution cipher (its own inverse)."
	transform := func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = w.Write(rot13(data))
		return err
	}
	p.Process = transform
	p.Unprocess = transform
	return p
}
