//go:build js && wasm

package webui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

var webBenchmarkSink any

func BenchmarkWebDisplayTextLargeCapped(b *testing.B) {
	data := []byte(strings.Repeat("deen\n", (pipeline.LargeDataThreshold/5)+1))
	b.ReportAllocs()
	for b.Loop() {
		text, capped := webTextDisplay(data, false)
		webBenchmarkSink = text
		webBenchmarkSink = capped
	}
}

func BenchmarkWebDisplayHexLargeCapped(b *testing.B) {
	data := bytes.Repeat([]byte{0x00, 0x41, 0xff, 0x20}, pipeline.LargeDataThreshold/4)
	b.ReportAllocs()
	for b.Loop() {
		text, capped := webHexDisplay(data, false)
		webBenchmarkSink = text
		webBenchmarkSink = capped
	}
}

func BenchmarkWebDisplayStringsLargeCapped(b *testing.B) {
	data := []byte(strings.Repeat("deen-perf\x00", pipeline.LargeDataThreshold/10))
	b.ReportAllocs()
	for b.Loop() {
		text, capped := webStringsDisplay(data, false)
		webBenchmarkSink = text
		webBenchmarkSink = capped
	}
}

func BenchmarkWebExamplesInitialFilter(b *testing.B) {
	examples := pipeline.BuiltinExamples()
	b.ReportAllocs()
	for b.Loop() {
		var matches int
		for _, example := range examples {
			if pipeline.ExampleMatches(example, "") {
				matches++
				webBenchmarkSink = exampleChainSummary(example.Steps)
				webBenchmarkSink = pipeline.DataMetadata(example.Source, 0).Summary()
			}
		}
		webBenchmarkSink = matches
	}
}

func BenchmarkWebExamplesResultPreviewCold(b *testing.B) {
	examples := pipeline.BuiltinExamples()
	b.ReportAllocs()
	for b.Loop() {
		for _, example := range examples {
			result, err := pipeline.ExampleResult(example)
			if err != nil {
				b.Fatal(err)
			}
			webBenchmarkSink = exampleDataText(example.Source)
			webBenchmarkSink = exampleDataText(result)
		}
	}
}

func BenchmarkWebExamplesResultPreviewCached(b *testing.B) {
	examples := pipeline.BuiltinExamples()
	examplePreviews = map[string]examplePreview{}
	for _, example := range examples {
		webBenchmarkSink = cachedExamplePreview(example)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		for _, example := range examples {
			preview := cachedExamplePreview(example)
			if preview.err != nil {
				b.Fatal(preview.err)
			}
			webBenchmarkSink = exampleDataText(example.Source)
			webBenchmarkSink = exampleDataText(preview.result)
		}
	}
}

func BenchmarkWebPluginCatalogPreparation(b *testing.B) {
	catalog := plugins.UICatalog()
	b.ReportAllocs()
	for b.Loop() {
		for _, info := range catalog {
			metaParts := []string{plugins.CategoryLabel(info.Category)}
			if info.CanDecode {
				metaParts = append(metaParts, "Encode and decode")
			} else {
				metaParts = append(metaParts, "Encode only")
			}
			if info.Label != info.Name {
				metaParts = append(metaParts, "Command: "+info.Name)
			}
			if len(info.Aliases) > 0 {
				metaParts = append(metaParts, "Aliases: "+strings.Join(info.Aliases, ", "))
			}
			webBenchmarkSink = strings.Join(metaParts, " · ")
			webBenchmarkSink = info.Description
			webBenchmarkSink = info.UseFor
		}
	}
}
