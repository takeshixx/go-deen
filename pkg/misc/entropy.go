package misc

import (
	"flag"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/takeshixx/deen/pkg/types"
)

type byteCount struct {
	b byte
	n int
}

func entropyBitsPerByte(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var counts [256]int
	for _, b := range data {
		counts[b]++
	}
	var entropy float64
	total := float64(len(data))
	for _, count := range counts {
		if count == 0 {
			continue
		}
		p := float64(count) / total
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// NewPluginEntropy creates a byte entropy and frequency analyzer.
func NewPluginEntropy() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "entropy"
	p.Aliases = []string{"freq"}
	p.Category = "misc"
	p.Description = "Analyze byte entropy, uniqueness and most common byte values."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		var counts [256]int
		var values []byteCount
		for _, b := range data {
			counts[b]++
		}
		for b, count := range counts {
			if count > 0 {
				values = append(values, byteCount{byte(b), count})
			}
		}
		sort.Slice(values, func(i, j int) bool {
			if values[i].n == values[j].n {
				return values[i].b < values[j].b
			}
			return values[i].n > values[j].n
		})
		fmt.Fprintf(w, "bytes: %d\nunique: %d\nentropy: %.4f bits/byte\n", len(data), len(values), entropyBitsPerByte(data))
		limit := min(10, len(values))
		for i := 0; i < limit; i++ {
			v := values[i]
			fmt.Fprintf(w, "0x%02x: %d %.2f%%\n", v.b, v.n, float64(v.n)*100/float64(max(1, len(data))))
		}
		return nil
	}
	return p
}
