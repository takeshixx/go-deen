package core

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestRunInspectWithArgsEmitsAgentJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runInspectWithArgs([]string{"-preview-bytes", "64", `{"ok":true}`}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	var resp inspectResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("inspect JSON did not decode: %v\n%s", err, stdout.String())
	}
	if resp.Version != 1 {
		t.Fatalf("version = %d, want 1", resp.Version)
	}
	if resp.Metadata.Bytes != len(`{"ok":true}`) {
		t.Fatalf("bytes = %d", resp.Metadata.Bytes)
	}
	if resp.Preview.Text != `{"ok":true}` {
		t.Fatalf("preview text = %q", resp.Preview.Text)
	}
	if !strings.Contains(resp.StructuredPreview, "JSON") {
		t.Fatalf("structured preview = %q", resp.StructuredPreview)
	}
	if !hasAgentSuggestion(resp.Suggestions, "json", false) {
		t.Fatalf("missing json suggestion: %#v", resp.Suggestions)
	}
}

func TestRunDetectWithArgsEmitsAutomatedChain(t *testing.T) {
	input := encodedGzipJSON(t, `{"ok":true}`)
	var stdout, stderr bytes.Buffer
	code := runDetectWithArgs([]string{input}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	var resp detectResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("detect JSON did not decode: %v\n%s", err, stdout.String())
	}
	if resp.Metadata.Bytes != len(input) {
		t.Fatalf("bytes = %d, want %d", resp.Metadata.Bytes, len(input))
	}
	if !hasAgentSuggestionChain(resp.Suggestions, []string{"url", "base64", "gzip", "json"}) {
		t.Fatalf("missing automated chain: %#v", resp.Suggestions)
	}
}

func encodedGzipJSON(t *testing.T, payload string) string {
	t.Helper()
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return url.QueryEscape(base64.StdEncoding.EncodeToString(gz.Bytes()))
}

func hasAgentSuggestion(suggestions []agentSuggestion, plugin string, unprocess bool) bool {
	for _, s := range suggestions {
		if s.Plugin == plugin && s.Unprocess == unprocess {
			return true
		}
	}
	return false
}

func hasAgentSuggestionChain(suggestions []agentSuggestion, plugins []string) bool {
	for _, s := range suggestions {
		if len(s.Steps) != len(plugins) {
			continue
		}
		matched := true
		for i, plugin := range plugins {
			if s.Steps[i].Plugin != plugin {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
