package core

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRunMCPWithArgsRejectsMissingServe(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runMCPWithArgs(nil, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestServeMCPInitializeAndListTools(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	)

	if got := jsonPath[string](t, out[0], "result", "serverInfo", "name"); got != "deen" {
		t.Fatalf("server name = %q", got)
	}
	tools := jsonPath[[]any](t, out[1], "result", "tools")
	if len(tools) < 4 {
		t.Fatalf("tool count = %d, want at least 4", len(tools))
	}
	if !mcpToolListContains(tools, "inspect") ||
		!mcpToolListContains(tools, "detect_next") ||
		!mcpToolListContains(tools, "read_result_range") {
		t.Fatalf("missing expected tools: %#v", tools)
	}
}

func TestServeMCPProtocolErrorsAndNotifications(t *testing.T) {
	out := serveMCPTranscriptExpectResponses(t, 3,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{bad json`,
		`{"jsonrpc":"2.0","id":"ping","method":"ping"}`,
		`{"jsonrpc":"2.0","id":"missing","method":"nope"}`,
	)
	if got := jsonPath[float64](t, out[0], "error", "code"); got != -32700 {
		t.Fatalf("parse error code = %v", got)
	}
	if got := jsonPath[map[string]any](t, out[1], "result"); len(got) != 0 {
		t.Fatalf("ping result = %#v", got)
	}
	if got := jsonPath[float64](t, out[2], "error", "code"); got != -32601 {
		t.Fatalf("missing method error code = %v", got)
	}
}

func TestServeMCPResources(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"list","method":"resources/list"}`,
		`{"jsonrpc":"2.0","id":"read","method":"resources/read","params":{"uri":"deen://plugins/base64"}}`,
		`{"jsonrpc":"2.0","id":"examples","method":"resources/read","params":{"uri":"deen://examples"}}`,
		`{"jsonrpc":"2.0","id":"example","method":"resources/read","params":{"uri":"deen://examples/compressed-webhook-payload"}}`,
	)
	resources := jsonPath[[]any](t, out[0], "result", "resources")
	if !mcpResourceListContains(resources, "deen://plugins") || !mcpResourceListContains(resources, "deen://plugins/base64") {
		t.Fatalf("missing expected resources: %#v", resources[:min(5, len(resources))])
	}
	text := jsonPath[string](t, out[1], "result", "contents", "0", "text")
	if !strings.Contains(text, `"Name": "base64"`) && !strings.Contains(text, `"name": "base64"`) {
		t.Fatalf("plugin resource text = %s", text)
	}
	examplesText := jsonPath[string](t, out[2], "result", "contents", "0", "text")
	if !strings.Contains(examplesText, "Compressed webhook payload") {
		t.Fatalf("examples resource text = %s", examplesText)
	}
	exampleText := jsonPath[string](t, out[3], "result", "contents", "0", "text")
	if !strings.Contains(exampleText, "URL-decode a copied request parameter") {
		t.Fatalf("example resource text = %s", exampleText)
	}
}

func TestServeMCPUnknownResource(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"read","method":"resources/read","params":{"uri":"deen://plugins/nope"}}`,
	)
	if got := jsonPath[float64](t, out[0], "error", "code"); got != -32602 {
		t.Fatalf("error code = %v", got)
	}
}

func TestServeMCPPrompts(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"list","method":"prompts/list"}`,
		`{"jsonrpc":"2.0","id":"get","method":"prompts/get","params":{"name":"decode_payload","arguments":{"goal":"decode a webhook"}}}`,
		`{"jsonrpc":"2.0","id":"triage","method":"prompts/get","params":{"name":"triage_unknown_data","arguments":{"data_hint":"from proxy logs"}}}`,
		`{"jsonrpc":"2.0","id":"binary","method":"prompts/get","params":{"name":"inspect_binary"}}`,
		`{"jsonrpc":"2.0","id":"chain","method":"prompts/get","params":{"name":"explain_chain","arguments":{"chain_json":"{\"version\":1,\"steps\":[]}"}}}`,
	)
	prompts := jsonPath[[]any](t, out[0], "result", "prompts")
	if !mcpNamedListContains(prompts, "decode_payload") {
		t.Fatalf("missing decode_payload prompt: %#v", prompts)
	}
	text := jsonPath[string](t, out[1], "result", "messages", "0", "content", "text")
	if !strings.Contains(text, "decode a webhook") {
		t.Fatalf("prompt text = %q", text)
	}
	if triage := jsonPath[string](t, out[2], "result", "messages", "0", "content", "text"); !strings.Contains(triage, "from proxy logs") {
		t.Fatalf("triage prompt text = %q", triage)
	}
	if binary := jsonPath[string](t, out[3], "result", "messages", "0", "content", "text"); !strings.Contains(binary, "Do not execute") {
		t.Fatalf("binary prompt text = %q", binary)
	}
	if chain := jsonPath[string](t, out[4], "result", "messages", "0", "content", "text"); !strings.Contains(chain, `"version":1`) {
		t.Fatalf("chain prompt text = %q", chain)
	}
}

func TestServeMCPUnknownPrompt(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"get","method":"prompts/get","params":{"name":"nope"}}`,
	)
	if got := jsonPath[float64](t, out[0], "error", "code"); got != -32602 {
		t.Fatalf("error code = %v", got)
	}
}

func TestServeMCPResultResource(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"transform","method":"tools/call","params":{"name":"transform","arguments":{"text":"test","plugin":"base64"}}}`,
		`{"jsonrpc":"2.0","id":"read","method":"resources/read","params":{"uri":"deen://results/result-1"}}`,
	)
	if got := jsonPath[string](t, out[0], "result", "structuredContent", "result_ref", "id"); got != "result-1" {
		t.Fatalf("result ref = %q", got)
	}
	text := jsonPath[string](t, out[1], "result", "contents", "0", "text")
	if !strings.Contains(text, `"ref": "result-1"`) {
		t.Fatalf("result resource = %s", text)
	}
}

func TestServeMCPInspectTool(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"inspect","method":"tools/call","params":{"name":"inspect","arguments":{"text":"{\"ok\":true}","preview_bytes":80}}}`,
	)
	if got := jsonPath[string](t, out[0], "result", "content", "0", "text"); got != "Inspection complete." {
		t.Fatalf("summary = %q", got)
	}
	if got := jsonPath[string](t, out[0], "result", "structuredContent", "preview", "text"); got != `{"ok":true}` {
		t.Fatalf("preview = %q", got)
	}
	suggestions := jsonPath[[]any](t, out[0], "result", "structuredContent", "suggestions")
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions")
	}
}

func TestServeMCPDetectNextTool(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"detect","method":"tools/call","params":{"name":"detect_next","arguments":{"text":"%7B%22ok%22%3Atrue%7D"}}}`,
	)
	suggestions := jsonPath[[]any](t, out[0], "result", "structuredContent", "suggestions")
	if !mcpSuggestionContains(suggestions, "url") {
		t.Fatalf("missing url suggestion: %#v", suggestions)
	}
}

func TestServeMCPTransformTool(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"transform","method":"tools/call","params":{"name":"transform","arguments":{"text":"test","plugin":"base64"}}}`,
	)
	if got := jsonPath[string](t, out[0], "result", "structuredContent", "preview", "text"); got != "dGVzdA==" {
		t.Fatalf("transform preview = %q", got)
	}
	if got := jsonPath[string](t, out[0], "result", "structuredContent", "result_ref", "id"); got == "" {
		t.Fatal("missing result ref")
	}
}

func TestServeMCPTransformErrors(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"missing","method":"tools/call","params":{"name":"transform","arguments":{"text":"test"}}}`,
		`{"jsonrpc":"2.0","id":"bad64","method":"tools/call","params":{"name":"transform","arguments":{"base64":"%%%","plugin":"base64"}}}`,
	)
	if got := jsonPath[bool](t, out[0], "result", "isError"); !got {
		t.Fatal("missing plugin should be tool error")
	}
	if got := jsonPath[bool](t, out[1], "result", "isError"); !got {
		t.Fatal("bad base64 should be tool error")
	}
}

func TestServeMCPTransformBase64Input(t *testing.T) {
	input := base64.StdEncoding.EncodeToString([]byte("test"))
	out := serveMCPTranscript(t,
		fmt.Sprintf(`{"jsonrpc":"2.0","id":"transform","method":"tools/call","params":{"name":"transform","arguments":{"base64":%q,"plugin":"base64"}}}`, input),
	)
	if got := jsonPath[string](t, out[0], "result", "structuredContent", "preview", "text"); got != "dGVzdA==" {
		t.Fatalf("transform preview = %q", got)
	}
}

func TestServeMCPRunChainTool(t *testing.T) {
	chain := `{"version":1,"steps":[{"plugin":"url","unprocess":true},{"plugin":"json"}]}`
	out := serveMCPTranscript(t,
		fmt.Sprintf(`{"jsonrpc":"2.0","id":"chain","method":"tools/call","params":{"name":"run_chain","arguments":{"text":"%%7B%%22ok%%22%%3Atrue%%7D","chain_json":%q}}}`, chain),
	)
	if got := jsonPath[string](t, out[0], "result", "content", "0", "text"); got != "Chain complete." {
		t.Fatalf("summary = %q", got)
	}
	if preview := jsonPath[string](t, out[0], "result", "structuredContent", "preview", "text"); !strings.Contains(preview, `"ok": true`) {
		t.Fatalf("chain preview = %q", preview)
	}
}

func TestServeMCPPluginTools(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"list","method":"tools/call","params":{"name":"list_plugins","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":"search","method":"tools/call","params":{"name":"search_plugins","arguments":{"query":"base64"}}}`,
	)
	if plugins := jsonPath[[]any](t, out[0], "result", "structuredContent"); len(plugins) == 0 {
		t.Fatal("empty plugin list")
	}
	matches := jsonPath[[]any](t, out[1], "result", "structuredContent")
	if !mcpNamedListContains(matches, "base64") {
		t.Fatalf("missing base64 search result: %#v", matches)
	}
}

func TestServeMCPReadResultRangeTool(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"transform","method":"tools/call","params":{"name":"transform","arguments":{"text":"test","plugin":"base64"}}}`,
		`{"jsonrpc":"2.0","id":"range","method":"tools/call","params":{"name":"read_result_range","arguments":{"ref":"result-1","offset":1,"length":3}}}`,
	)
	if got := jsonPath[string](t, out[1], "result", "content", "0", "text"); got != "Result range read." {
		t.Fatalf("summary = %q", got)
	}
	if got := jsonPath[string](t, out[1], "result", "structuredContent", "preview", "text"); got != "GVz" {
		t.Fatalf("range preview = %q", got)
	}
	if got := jsonPath[float64](t, out[1], "result", "structuredContent", "remaining"); got != 4 {
		t.Fatalf("remaining = %v, want 4", got)
	}
}

func TestServeMCPReadResultRangeEdges(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"transform","method":"tools/call","params":{"name":"transform","arguments":{"text":"test","plugin":"base64"}}}`,
		`{"jsonrpc":"2.0","id":"past-end","method":"tools/call","params":{"name":"read_result_range","arguments":{"ref":"result-1","offset":99,"length":3}}}`,
		`{"jsonrpc":"2.0","id":"default-len","method":"tools/call","params":{"name":"read_result_range","arguments":{"ref":"result-1"}}}`,
		`{"jsonrpc":"2.0","id":"missing","method":"tools/call","params":{"name":"read_result_range","arguments":{"ref":"nope"}}}`,
		`{"jsonrpc":"2.0","id":"negative","method":"tools/call","params":{"name":"read_result_range","arguments":{"ref":"result-1","offset":-1}}}`,
	)
	if got := jsonPath[float64](t, out[1], "result", "structuredContent", "length"); got != 0 {
		t.Fatalf("past-end length = %v", got)
	}
	if got := jsonPath[float64](t, out[2], "result", "structuredContent", "length"); got != 8 {
		t.Fatalf("default length = %v", got)
	}
	if got := jsonPath[bool](t, out[3], "result", "isError"); !got {
		t.Fatal("unknown ref should be tool error")
	}
	if got := jsonPath[bool](t, out[4], "result", "isError"); !got {
		t.Fatal("negative offset should be tool error")
	}
}

func TestServeMCPUnknownTool(t *testing.T) {
	out := serveMCPTranscript(t,
		`{"jsonrpc":"2.0","id":"unknown","method":"tools/call","params":{"name":"nope","arguments":{}}}`,
	)
	if got := jsonPath[float64](t, out[0], "error", "code"); got != -32602 {
		t.Fatalf("error code = %v", got)
	}
}

func serveMCPTranscript(t *testing.T, lines ...string) []map[string]any {
	return serveMCPTranscriptExpectResponses(t, len(lines), lines...)
}

func serveMCPTranscriptExpectResponses(t *testing.T, wantResponses int, lines ...string) []map[string]any {
	t.Helper()
	var in strings.Builder
	for _, line := range lines {
		in.WriteString(line)
		in.WriteByte('\n')
	}
	var out bytes.Buffer
	if err := serveMCP(strings.NewReader(in.String()), &out); err != nil {
		t.Fatal(err)
	}
	var responses []map[string]any
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		var msg map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			t.Fatalf("response JSON: %v\n%s", err, scanner.Text())
		}
		responses = append(responses, msg)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(responses) != wantResponses {
		t.Fatalf("responses = %d, want %d\n%s", len(responses), wantResponses, out.String())
	}
	return responses
}

func mcpToolListContains(tools []any, name string) bool {
	return mcpNamedListContains(tools, name)
}

func mcpNamedListContains(items []any, name string) bool {
	for _, item := range items {
		m, ok := item.(map[string]any)
		if ok && (m["name"] == name || m["Name"] == name) {
			return true
		}
	}
	return false
}

func mcpSuggestionContains(items []any, plugin string) bool {
	for _, item := range items {
		m, ok := item.(map[string]any)
		if ok && m["plugin"] == plugin {
			return true
		}
	}
	return false
}

func mcpResourceListContains(resources []any, uri string) bool {
	for _, item := range resources {
		resource, ok := item.(map[string]any)
		if ok && resource["uri"] == uri {
			return true
		}
	}
	return false
}

func jsonPath[T any](t *testing.T, msg map[string]any, path ...string) T {
	t.Helper()
	var cur any = msg
	for _, part := range path {
		if list, ok := cur.([]any); ok {
			var index int
			if _, err := fmt.Sscanf(part, "%d", &index); err != nil || index < 0 || index >= len(list) {
				t.Fatalf("bad list index %q in %#v", part, cur)
			}
			cur = list[index]
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("path %v hit non-object %#v", path, cur)
		}
		cur = m[part]
	}
	got, ok := cur.(T)
	if !ok {
		t.Fatalf("path %v type = %T, want requested type", path, cur)
	}
	return got
}
