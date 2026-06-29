package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/pkg/helpers"
)

const defaultAgentPreviewBytes = 2048

type agentMetadata struct {
	Bytes      int     `json:"bytes"`
	Lines      int     `json:"lines"`
	UTF8       bool    `json:"utf8"`
	Encoding   string  `json:"encoding"`
	BOM        string  `json:"bom,omitempty"`
	Printable  int     `json:"printable_percent"`
	Entropy    float64 `json:"entropy_bits_per_byte"`
	InputBytes int     `json:"input_bytes,omitempty"`
	Sampled    bool    `json:"sampled"`
}

type agentPreview struct {
	Text      string `json:"text,omitempty"`
	Hex       string `json:"hex,omitempty"`
	Base64    string `json:"base64,omitempty"`
	Truncated bool   `json:"truncated"`
}

type inspectResponse struct {
	Version           int               `json:"version"`
	SHA256            string            `json:"sha256"`
	MIME              string            `json:"mime"`
	Metadata          agentMetadata     `json:"metadata"`
	Preview           agentPreview      `json:"preview"`
	ResultRef         *agentResultRef   `json:"result_ref,omitempty"`
	StructuredPreview string            `json:"structured_preview,omitempty"`
	Suggestions       []agentSuggestion `json:"suggestions,omitempty"`
}

type detectResponse struct {
	Version     int               `json:"version"`
	SHA256      string            `json:"sha256"`
	Metadata    agentMetadata     `json:"metadata"`
	Suggestions []agentSuggestion `json:"suggestions"`
}

type agentSuggestion struct {
	Label      string                `json:"label"`
	Reason     string                `json:"reason"`
	Confidence int                   `json:"confidence,omitempty"`
	Preview    string                `json:"preview,omitempty"`
	Plugin     string                `json:"plugin"`
	Unprocess  bool                  `json:"unprocess"`
	Options    map[string]string     `json:"options,omitempty"`
	Steps      []agentSuggestionStep `json:"steps"`
}

type agentSuggestionStep struct {
	Plugin    string            `json:"plugin"`
	Unprocess bool              `json:"unprocess"`
	Options   map[string]string `json:"options,omitempty"`
}

type agentResultRef struct {
	ID          string `json:"id"`
	Bytes       int    `json:"bytes"`
	SHA256      string `json:"sha256"`
	Description string `json:"description,omitempty"`
}

func runInspect() int {
	return runInspectWithArgs(helpers.RemoveBeforeSubcommand(os.Args, "inspect"), os.Stdin, os.Stdout, os.Stderr)
}

func runDetect() int {
	return runDetectWithArgs(helpers.RemoveBeforeSubcommand(os.Args, "detect"), os.Stdin, os.Stdout, os.Stderr)
}

func runInspectWithArgs(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := agentFlagSet("inspect", stderr)
	previewBytes := fs.Int("preview-bytes", defaultAgentPreviewBytes, "maximum preview bytes to include")
	file := fs.String("file", "", "read input from file")
	fs.Parse(args)

	data, err := readAgentInput(*file, fs.Args(), stdin)
	if err != nil {
		fmt.Fprintln(stderr, "deen: inspect:", err)
		return 1
	}

	return writeAgentJSON(stdout, stderr, inspectData(data, *previewBytes))
}

func runDetectWithArgs(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := agentFlagSet("detect", stderr)
	file := fs.String("file", "", "read input from file")
	fs.Parse(args)

	data, err := readAgentInput(*file, fs.Args(), stdin)
	if err != nil {
		fmt.Fprintln(stderr, "deen: detect:", err)
		return 1
	}

	return writeAgentJSON(stdout, stderr, detectData(data))
}

func inspectData(data []byte, previewBytes int) inspectResponse {
	resp := inspectResponse{
		Version:     1,
		SHA256:      sha256Hex(data),
		MIME:        http.DetectContentType(sampleForSniffing(data)),
		Metadata:    agentMetadataFromPipeline(pipeline.DataMetadata(data, 0)),
		Preview:     agentPreviewForData(data, previewBytes),
		Suggestions: agentSuggestions(pipeline.Suggestions(data)),
	}
	if preview, ok := pipeline.StructuredPreview(data); ok {
		resp.StructuredPreview = capString(preview, max(0, previewBytes))
	}
	return resp
}

func detectData(data []byte) detectResponse {
	return detectResponse{
		Version:     1,
		SHA256:      sha256Hex(data),
		Metadata:    agentMetadataFromPipeline(pipeline.DataMetadata(data, 0)),
		Suggestions: agentSuggestions(pipeline.Suggestions(data)),
	}
}

func agentFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage of %s:\n\n", name)
		fmt.Fprintf(stderr, "Emit agent-friendly JSON for local data analysis.\n\n")
		fmt.Fprintf(stderr, "Examples:\n")
		fmt.Fprintf(stderr, "  deen %s sample\n", name)
		fmt.Fprintf(stderr, "  printf data | deen %s\n", name)
		fmt.Fprintf(stderr, "  deen %s -file payload.bin\n\n", name)
		fs.PrintDefaults()
	}
	return fs
}

func readAgentInput(path string, args []string, stdin io.Reader) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	if len(args) > 0 {
		return []byte(strings.Join(args, " ")), nil
	}
	return io.ReadAll(stdin)
}

func agentMetadataFromPipeline(m pipeline.Metadata) agentMetadata {
	bom := m.BOM
	if bom == "none" {
		bom = ""
	}
	return agentMetadata{
		Bytes:      m.Bytes,
		Lines:      m.Lines,
		UTF8:       m.UTF8,
		Encoding:   m.Encoding,
		BOM:        bom,
		Printable:  m.Printable,
		Entropy:    m.Entropy,
		InputBytes: m.InputBytes,
		Sampled:    m.Sampled,
	}
}

func agentPreviewForData(data []byte, limit int) agentPreview {
	if limit < 0 {
		limit = 0
	}
	sample := data
	truncated := false
	if len(sample) > limit {
		sample = sample[:limit]
		truncated = true
	}
	preview := agentPreview{Truncated: truncated}
	if utf8.Valid(sample) && !pipeline.IsBinaryData(sample) {
		preview.Text = string(sample)
		return preview
	}
	preview.Hex = hex.Dump(sample)
	preview.Base64 = base64.StdEncoding.EncodeToString(sample)
	return preview
}

func agentSuggestions(suggestions []pipeline.Suggestion) []agentSuggestion {
	out := make([]agentSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		steps := s.Steps
		if len(steps) == 0 && s.Plugin != "" {
			steps = []pipeline.SuggestionStep{{
				Plugin:    s.Plugin,
				Unprocess: s.Unprocess,
				Options:   s.Options,
			}}
		}
		out = append(out, agentSuggestion{
			Label:      s.Label,
			Reason:     s.Reason,
			Confidence: s.Confidence,
			Preview:    s.Preview,
			Plugin:     s.Plugin,
			Unprocess:  s.Unprocess,
			Options:    cloneStringMap(s.Options),
			Steps:      agentSuggestionSteps(steps),
		})
	}
	return out
}

func agentSuggestionSteps(steps []pipeline.SuggestionStep) []agentSuggestionStep {
	out := make([]agentSuggestionStep, 0, len(steps))
	for _, step := range steps {
		out = append(out, agentSuggestionStep{
			Plugin:    step.Plugin,
			Unprocess: step.Unprocess,
			Options:   cloneStringMap(step.Options),
		})
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func writeAgentJSON(stdout, stderr io.Writer, v any) int {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(stderr, "deen:", err)
		return 1
	}
	return 0
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sampleForSniffing(data []byte) []byte {
	if len(data) >= 512 {
		return data[:512]
	}
	return data
}

func capString(s string, limit int) string {
	if limit == 0 {
		return ""
	}
	if len(s) <= limit {
		return s
	}
	return s[:safeStringCut(s, limit)] + "\n... truncated ..."
}

func safeStringCut(s string, limit int) int {
	if limit >= len(s) {
		return len(s)
	}
	for limit > 0 && !utf8.RuneStart(s[limit]) {
		limit--
	}
	return limit
}
