package compressions

import (
	"bytes"
	"testing"

	"github.com/takeshixx/deen/pkg/types"
)

// TestCompressionsBinarySafe verifies every compression plugin round-trips
// arbitrary binary data (all 256 byte values) unchanged.
func TestCompressionsBinarySafe(t *testing.T) {
	plugins := map[string]*types.DeenPlugin{
		"flate":  NewPluginFlate(),
		"gzip":   NewPluginGzip(),
		"zlib":   NewPluginZlib(),
		"bzip2":  NewPluginBzip2(),
		"lzma":   NewPluginLZMA(),
		"lzma2":  NewPluginLZMA2(),
		"lzw":    NewPluginLzw(),
		"brotli": NewPluginBrotli(),
		"zstd":   NewPluginZstd(),
	}
	input := make([]byte, 256)
	for i := range input {
		input[i] = byte(i)
	}
	for name, p := range plugins {
		compressed, err := transform(p.Process, p.RegisterFlags, input)
		if err != nil {
			t.Errorf("%s: compress failed: %s", name, err)
			continue
		}
		out, err := transform(p.Unprocess, p.RegisterFlags, compressed)
		if err != nil {
			t.Errorf("%s: decompress failed: %s", name, err)
			continue
		}
		if !bytes.Equal(out, input) {
			t.Errorf("%s: binary round-trip mismatch", name)
		}
	}
}
