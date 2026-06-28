package pipeline

import (
	"bytes"
	"strings"
	"testing"
)

var benchmarkSink any

func BenchmarkDisplayTextSmall(b *testing.B) {
	data := []byte(strings.Repeat("deen\n", 1024))
	b.ReportAllocs()
	for b.Loop() {
		text, capped := TextDisplay(data)
		benchmarkSink = text
		benchmarkSink = capped
	}
}

func BenchmarkDisplayTextLargeCapped(b *testing.B) {
	data := []byte(strings.Repeat("deen\n", (LargeDataThreshold/5)+1))
	b.ReportAllocs()
	for b.Loop() {
		text, capped := TextDisplay(data)
		benchmarkSink = text
		benchmarkSink = capped
	}
}

func BenchmarkDisplayHexLargeCapped(b *testing.B) {
	data := bytes.Repeat([]byte{0x00, 0x41, 0xff, 0x20}, LargeDataThreshold/4)
	b.ReportAllocs()
	for b.Loop() {
		text, capped := HexDisplay(data)
		benchmarkSink = text
		benchmarkSink = capped
	}
}

func BenchmarkDisplayStringsLargeCapped(b *testing.B) {
	data := []byte(strings.Repeat("deen-perf\x00", LargeDataThreshold/10))
	b.ReportAllocs()
	for b.Loop() {
		text, capped := StringsDisplay(data)
		benchmarkSink = text
		benchmarkSink = capped
	}
}

func BenchmarkDataMetadataLarge(b *testing.B) {
	data := bytes.Repeat([]byte{0x00, 0x41, 0xff, 0x20}, LargeDataThreshold/4)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkSink = DataMetadata(data, 0)
	}
}

func BenchmarkHighlightedPreviewJSON(b *testing.B) {
	data := []byte(`{"token":"abc.def.ghi","nested":{"enabled":true,"items":[1,2,3],"name":"deen"}}`)
	b.ReportAllocs()
	for b.Loop() {
		preview, spans, ok := HighlightedPreview(data)
		benchmarkSink = preview
		benchmarkSink = spans
		benchmarkSink = ok
	}
}

func BenchmarkBuiltinExamplesResult(b *testing.B) {
	examples := BuiltinExamples()
	b.ReportAllocs()
	for b.Loop() {
		for _, example := range examples {
			result, err := ExampleResult(example)
			if err != nil {
				b.Fatal(err)
			}
			benchmarkSink = result
		}
	}
}

func BenchmarkPipelineTenStepRefresh(b *testing.B) {
	source := []byte(strings.Repeat("deen", 1024))
	b.ReportAllocs()
	for b.Loop() {
		p := New()
		p.SetSource(source)
		for i := 0; i < 10; i++ {
			p.AddStep("base64", i%2 == 1)
		}
		benchmarkSink = p.Result()
	}
}
