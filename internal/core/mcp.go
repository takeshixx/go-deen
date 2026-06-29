package core

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
	"github.com/takeshixx/deen/pkg/helpers"
)

const mcpProtocolVersion = "2025-06-18"
const defaultMCPRangeLength = 4096

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type mcpReadResourceParams struct {
	URI string `json:"uri"`
}

type mcpGetPromptParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

type mcpDataArgs struct {
	Text         string `json:"text,omitempty"`
	Base64       string `json:"base64,omitempty"`
	PreviewBytes int    `json:"preview_bytes,omitempty"`
}

type mcpTransformArgs struct {
	Text      string            `json:"text,omitempty"`
	Base64    string            `json:"base64,omitempty"`
	Plugin    string            `json:"plugin"`
	Unprocess bool              `json:"unprocess,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
}

type mcpRunChainArgs struct {
	Text      string `json:"text,omitempty"`
	Base64    string `json:"base64,omitempty"`
	ChainJSON string `json:"chain_json"`
}

type mcpSearchArgs struct {
	Query string `json:"query,omitempty"`
}

type mcpReadRangeArgs struct {
	Ref    string `json:"ref"`
	Offset int    `json:"offset,omitempty"`
	Length int    `json:"length,omitempty"`
}

type mcpStoredResult struct {
	id   string
	data []byte
	sum  string
}

type mcpSession struct {
	results map[string]mcpStoredResult
	nextID  int
}

func runMCP() int {
	return runMCPWithArgs(helpers.RemoveBeforeSubcommand(os.Args, "mcp"), os.Stdin, os.Stdout, os.Stderr)
}

func runMCPWithArgs(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage of mcp:")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "  deen mcp serve")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Run a stdio MCP server for local, agent-friendly deen tools.")
	}
	fs.Parse(args)
	if fs.NArg() != 1 || fs.Arg(0) != "serve" {
		fs.Usage()
		return 2
	}
	if err := serveMCP(stdin, stdout); err != nil {
		fmt.Fprintln(stderr, "deen: mcp:", err)
		return 1
	}
	return 0
}

func serveMCP(stdin io.Reader, stdout io.Writer) error {
	session := &mcpSession{results: map[string]mcpStoredResult{}}
	scanner := bufio.NewScanner(stdin)
	// MCP messages are small JSON-RPC requests; allow comfortably larger tool
	// payloads without turning the scanner into an unbounded read.
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	enc := json.NewEncoder(stdout)
	enc.SetEscapeHTML(false)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		resp := session.handleMCPLine(line)
		if resp == nil {
			continue
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *mcpSession) handleMCPLine(line []byte) *mcpResponse {
	var req mcpRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return &mcpResponse{JSONRPC: "2.0", Error: &mcpError{Code: -32700, Message: "parse error"}}
	}
	if len(req.ID) == 0 {
		return nil
	}
	result, err := s.handleMCPRequest(req)
	if err != nil {
		return &mcpResponse{JSONRPC: "2.0", ID: req.ID, Error: err}
	}
	return &mcpResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func (s *mcpSession) handleMCPRequest(req mcpRequest) (any, *mcpError) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "deen",
				"version": Version(),
			},
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": mcpTools()}, nil
	case "tools/call":
		var params mcpCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, mcpInvalidParams("invalid tools/call params")
		}
		return s.callMCPTool(params)
	case "resources/list":
		return map[string]any{"resources": s.mcpResources()}, nil
	case "resources/read":
		var params mcpReadResourceParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, mcpInvalidParams("invalid resources/read params")
		}
		return s.readMCPResource(params.URI)
	case "prompts/list":
		return map[string]any{"prompts": mcpPrompts()}, nil
	case "prompts/get":
		var params mcpGetPromptParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, mcpInvalidParams("invalid prompts/get params")
		}
		return getMCPPrompt(params)
	default:
		return nil, &mcpError{Code: -32601, Message: "method not found"}
	}
}

func (s *mcpSession) mcpResources() []map[string]string {
	resources := []map[string]string{
		{
			"uri":         "deen://plugins",
			"name":        "Plugin catalog",
			"description": "All deen plugins with labels, categories, aliases, descriptions, use cases, examples, and decode support.",
			"mimeType":    "application/json",
		},
		{
			"uri":         "deen://examples",
			"name":        "Built-in examples",
			"description": "Runnable deen example inputs and transform chains.",
			"mimeType":    "application/json",
		},
	}
	for _, info := range plugins.UICatalog() {
		resources = append(resources, map[string]string{
			"uri":         "deen://plugins/" + info.Name,
			"name":        "Plugin: " + info.Label,
			"description": info.Description,
			"mimeType":    "application/json",
		})
	}
	for _, example := range pipeline.BuiltinExamples() {
		resources = append(resources, map[string]string{
			"uri":         "deen://examples/" + mcpSlug(example.Name),
			"name":        "Example: " + example.Name,
			"description": example.Description,
			"mimeType":    "application/json",
		})
	}
	for id, stored := range s.results {
		resources = append(resources, map[string]string{
			"uri":         "deen://results/" + id,
			"name":        "MCP result " + id,
			"description": fmt.Sprintf("%d bytes, sha256 %s", len(stored.data), stored.sum),
			"mimeType":    "application/octet-stream",
		})
	}
	return resources
}

func (s *mcpSession) readMCPResource(uri string) (any, *mcpError) {
	switch {
	case uri == "deen://plugins":
		return mcpResourceText(uri, plugins.UICatalog()), nil
	case strings.HasPrefix(uri, "deen://plugins/"):
		name := strings.TrimPrefix(uri, "deen://plugins/")
		for _, info := range plugins.UICatalog() {
			if info.Name == name {
				return mcpResourceText(uri, info), nil
			}
		}
		return nil, mcpInvalidParams("unknown plugin resource: " + uri)
	case uri == "deen://examples":
		return mcpResourceText(uri, pipeline.BuiltinExamples()), nil
	case strings.HasPrefix(uri, "deen://examples/"):
		slug := strings.TrimPrefix(uri, "deen://examples/")
		for _, example := range pipeline.BuiltinExamples() {
			if mcpSlug(example.Name) == slug {
				return mcpResourceText(uri, example), nil
			}
		}
		return nil, mcpInvalidParams("unknown example resource: " + uri)
	case strings.HasPrefix(uri, "deen://results/"):
		id := strings.TrimPrefix(uri, "deen://results/")
		stored, ok := s.results[id]
		if !ok {
			return nil, mcpInvalidParams("unknown result resource: " + uri)
		}
		return map[string]any{
			"contents": []map[string]any{{
				"uri":      uri,
				"mimeType": "application/json",
				"text": mustJSON(map[string]any{
					"ref":     stored.id,
					"bytes":   len(stored.data),
					"sha256":  stored.sum,
					"preview": agentPreviewForData(stored.data, defaultAgentPreviewBytes),
				}),
			}},
		}, nil
	default:
		return nil, mcpInvalidParams("unknown resource: " + uri)
	}
}

func mcpResourceText(uri string, v any) map[string]any {
	return map[string]any{
		"contents": []map[string]string{{
			"uri":      uri,
			"mimeType": "application/json",
			"text":     mustJSON(v),
		}},
	}
}

func mcpPrompts() []map[string]any {
	return []map[string]any{
		{
			"name":        "triage_unknown_data",
			"description": "Inspect unknown local data, summarize metadata, and propose safe decode/inspection next steps.",
			"arguments": []map[string]any{{
				"name":        "data_hint",
				"description": "Optional context about where the data came from.",
				"required":    false,
			}},
		},
		{
			"name":        "decode_payload",
			"description": "Use detect_next and deen transforms to decode a layered payload while keeping outputs bounded.",
			"arguments": []map[string]any{{
				"name":        "goal",
				"description": "What the decoded payload should reveal.",
				"required":    false,
			}},
		},
		{
			"name":        "inspect_binary",
			"description": "Inspect an executable or binary blob with metadata, magic, entropy, strings, and binary structure tools.",
			"arguments":   []map[string]any{},
		},
		{
			"name":        "explain_chain",
			"description": "Explain what a deen transform chain does and what input/output shape it expects.",
			"arguments": []map[string]any{{
				"name":        "chain_json",
				"description": "Optional deen chain JSON to explain.",
				"required":    false,
			}},
		},
	}
}

func getMCPPrompt(params mcpGetPromptParams) (any, *mcpError) {
	text, ok := mcpPromptText(params.Name, params.Arguments)
	if !ok {
		return nil, mcpInvalidParams("unknown prompt: " + params.Name)
	}
	return map[string]any{
		"description": params.Name,
		"messages": []map[string]any{{
			"role": "user",
			"content": map[string]string{
				"type": "text",
				"text": text,
			},
		}},
	}, nil
}

func mcpPromptText(name string, args map[string]string) (string, bool) {
	switch name {
	case "triage_unknown_data":
		hint := strings.TrimSpace(args["data_hint"])
		if hint != "" {
			hint = " Context: " + hint
		}
		return "Use deen inspect first, then detect_next if the input appears encoded, compressed, serialized, certificate-like, token-like, or binary." + hint + " Keep outputs bounded and cite SHA-256 plus result refs when available.", true
	case "decode_payload":
		goal := strings.TrimSpace(args["goal"])
		if goal == "" {
			goal = "recover the most readable structured representation"
		}
		return "Use deen detect_next to find candidate chains, apply the most plausible local chain with transform or run_chain, and read result ranges only when previews are insufficient. Goal: " + goal + ".", true
	case "inspect_binary":
		return "Use deen inspect to collect metadata and magic. If executable signatures are present, use bininspect via transform. Use entropy and bounded result ranges for suspicious binary regions. Do not execute the sample.", true
	case "explain_chain":
		chain := strings.TrimSpace(args["chain_json"])
		if chain == "" {
			return "Explain the deen chain step by step: each plugin, direction, options, expected input shape, output shape, and safe ways to validate the result.", true
		}
		return "Explain this deen chain step by step, including each plugin, direction, options, expected input shape, output shape, and safe validation hints:\n\n" + chain, true
	default:
		return "", false
	}
}

func mcpTools() []mcpTool {
	dataSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text":          map[string]any{"type": "string", "description": "Input text to analyze."},
			"base64":        map[string]any{"type": "string", "description": "Base64-encoded input bytes."},
			"preview_bytes": map[string]any{"type": "integer", "minimum": 0, "description": "Maximum bytes/characters in previews."},
		},
	}
	return []mcpTool{
		{
			Name:        "inspect",
			Description: "Inspect local data and return SHA-256, MIME, metadata, capped preview, structured preview, and likely deen transforms.",
			InputSchema: dataSchema,
		},
		{
			Name:        "detect_next",
			Description: "Suggest likely one-step and multi-step local decode or inspection chains for unknown encoded, compressed, serialized, token, certificate, or binary data.",
			InputSchema: dataSchema,
		},
		{
			Name:        "transform",
			Description: "Run one deen plugin transform locally on text or base64 input. Use unprocess=true for decode/reverse mode.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"plugin"},
				"properties": map[string]any{
					"text":      map[string]any{"type": "string"},
					"base64":    map[string]any{"type": "string"},
					"plugin":    map[string]any{"type": "string"},
					"unprocess": map[string]any{"type": "boolean"},
					"options":   map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}},
				},
			},
		},
		{
			Name:        "run_chain",
			Description: "Run a saved deen chain JSON recipe locally against text or base64 input.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"chain_json"},
				"properties": map[string]any{
					"text":       map[string]any{"type": "string"},
					"base64":     map[string]any{"type": "string"},
					"chain_json": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "list_plugins",
			Description: "List all available deen plugins with descriptions, categories, aliases, and decode support.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			Name:        "search_plugins",
			Description: "Search deen plugins by name, alias, category, description, and use-case text.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "read_result_range",
			Description: "Read a bounded byte range from an MCP result_ref returned by inspect, transform, or run_chain.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"ref"},
				"properties": map[string]any{
					"ref":    map[string]any{"type": "string", "description": "Result reference id."},
					"offset": map[string]any{"type": "integer", "minimum": 0},
					"length": map[string]any{"type": "integer", "minimum": 1},
				},
			},
		},
	}
}

func (s *mcpSession) callMCPTool(params mcpCallParams) (any, *mcpError) {
	switch params.Name {
	case "inspect":
		var args mcpDataArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid inspect arguments")
		}
		data, err := mcpInputBytes(args.Text, args.Base64)
		if err != nil {
			return nil, mcpInvalidParams(err.Error())
		}
		previewBytes := args.PreviewBytes
		if previewBytes == 0 {
			previewBytes = defaultAgentPreviewBytes
		}
		result := inspectData(data, previewBytes)
		s.attachResultRef(&result, data, "input data")
		return mcpToolResult("Inspection complete.", result, false), nil
	case "detect_next":
		var args mcpDataArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid detect_next arguments")
		}
		data, err := mcpInputBytes(args.Text, args.Base64)
		if err != nil {
			return nil, mcpInvalidParams(err.Error())
		}
		return mcpToolResult("Detection complete.", detectData(data), false), nil
	case "transform":
		var args mcpTransformArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid transform arguments")
		}
		data, err := mcpTransform(args)
		if err != nil {
			return mcpToolResult(err.Error(), map[string]any{"error": err.Error()}, true), nil
		}
		result := inspectData(data, defaultAgentPreviewBytes)
		s.attachResultRef(&result, data, "transform result")
		return mcpToolResult("Transform complete.", result, false), nil
	case "run_chain":
		var args mcpRunChainArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid run_chain arguments")
		}
		data, err := mcpRunChain(args)
		if err != nil {
			return mcpToolResult(err.Error(), map[string]any{"error": err.Error()}, true), nil
		}
		result := inspectData(data, defaultAgentPreviewBytes)
		s.attachResultRef(&result, data, "chain result")
		return mcpToolResult("Chain complete.", result, false), nil
	case "list_plugins":
		return mcpToolResult("Plugin list complete.", plugins.UICatalog(), false), nil
	case "search_plugins":
		var args mcpSearchArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid search_plugins arguments")
		}
		return mcpToolResult("Plugin search complete.", plugins.SearchUICatalog(args.Query), false), nil
	case "read_result_range":
		var args mcpReadRangeArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return nil, mcpInvalidParams("invalid read_result_range arguments")
		}
		result, err := s.readResultRange(args)
		if err != nil {
			return mcpToolResult(err.Error(), map[string]any{"error": err.Error()}, true), nil
		}
		return mcpToolResult("Result range read.", result, false), nil
	default:
		return nil, mcpInvalidParams("unknown tool: " + params.Name)
	}
}

func mcpToolResult(summary string, structured any, isError bool) map[string]any {
	return map[string]any{
		"content": []map[string]string{
			{"type": "text", "text": summary},
		},
		"structuredContent": structured,
		"isError":           isError,
	}
}

func mcpInputBytes(text, encoded string) ([]byte, error) {
	if encoded != "" {
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 input: %w", err)
		}
		return data, nil
	}
	return []byte(text), nil
}

func mcpTransform(args mcpTransformArgs) ([]byte, error) {
	if strings.TrimSpace(args.Plugin) == "" {
		return nil, errors.New("plugin is required")
	}
	data, err := mcpInputBytes(args.Text, args.Base64)
	if err != nil {
		return nil, err
	}
	pipe := pipeline.New()
	pipe.SetSourceOwned(data)
	pipe.AddStepWithOptions(args.Plugin, args.Unprocess, args.Options)
	if err := pipe.Err(0); err != nil {
		return nil, fmt.Errorf("%s: %w", args.Plugin, err)
	}
	return append([]byte(nil), pipe.Result()...), nil
}

func mcpRunChain(args mcpRunChainArgs) ([]byte, error) {
	if strings.TrimSpace(args.ChainJSON) == "" {
		return nil, errors.New("chain_json is required")
	}
	data, err := mcpInputBytes(args.Text, args.Base64)
	if err != nil {
		return nil, err
	}
	pipe := pipeline.New()
	if err := pipe.ImportJSON([]byte(args.ChainJSON)); err != nil {
		return nil, err
	}
	if args.Text != "" || args.Base64 != "" {
		pipe.SetSourceOwned(data)
	}
	if step, err := firstChainError(pipe); err != nil {
		return nil, fmt.Errorf("step %d (%s): %w", step+1, pipe.Steps()[step].Plugin, err)
	}
	return append([]byte(nil), pipe.Result()...), nil
}

func mcpInvalidParams(msg string) *mcpError {
	return &mcpError{Code: -32602, Message: msg}
}

func (s *mcpSession) attachResultRef(resp *inspectResponse, data []byte, description string) {
	stored := s.storeResult(data)
	resp.ResultRef = &agentResultRef{
		ID:          stored.id,
		Bytes:       len(stored.data),
		SHA256:      stored.sum,
		Description: description,
	}
}

func (s *mcpSession) storeResult(data []byte) mcpStoredResult {
	s.nextID++
	id := "result-" + strconv.Itoa(s.nextID)
	stored := mcpStoredResult{
		id:   id,
		data: append([]byte(nil), data...),
		sum:  sha256Hex(data),
	}
	s.results[id] = stored
	return stored
}

func (s *mcpSession) readResultRange(args mcpReadRangeArgs) (map[string]any, error) {
	stored, ok := s.results[args.Ref]
	if !ok {
		return nil, fmt.Errorf("unknown result ref %q", args.Ref)
	}
	if args.Offset < 0 {
		return nil, errors.New("offset must be >= 0")
	}
	length := args.Length
	if length <= 0 {
		length = defaultMCPRangeLength
	}
	if args.Offset > len(stored.data) {
		args.Offset = len(stored.data)
	}
	end := min(len(stored.data), args.Offset+length)
	chunk := stored.data[args.Offset:end]
	return map[string]any{
		"ref":       stored.id,
		"sha256":    stored.sum,
		"offset":    args.Offset,
		"length":    len(chunk),
		"end":       end,
		"total":     len(stored.data),
		"remaining": len(stored.data) - end,
		"preview":   agentPreviewForData(chunk, len(chunk)),
	}, nil
}

func mustJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return `{"error":"failed to encode resource"}`
	}
	return string(data)
}

func mcpSlug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
