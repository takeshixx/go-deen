package core

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/takeshixx/deen/internal/pipeline"
)

func TestSelectChainInputUsesArgs(t *testing.T) {
	r, cleanup, err := selectChainInput("", false, []string{"hello", "world"}, strings.NewReader("stdin"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "hello world" {
		t.Fatalf("chain input = %q", got)
	}
}

func TestSelectChainInputPrecedence(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(inputPath, []byte("from-file"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, cleanup, err := selectChainInput(inputPath, true, []string{"from", "args"}, strings.NewReader("from-stdin"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "from-file" {
		t.Fatalf("chain input = %q, want file input", got)
	}
}

func TestSelectChainInputUsesStdin(t *testing.T) {
	r, cleanup, err := selectChainInput("", true, nil, strings.NewReader("from-stdin"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "from-stdin" {
		t.Fatalf("chain input = %q, want stdin input", got)
	}
}

func TestRunChainWithSavedSource(t *testing.T) {
	chainPath := writeTestChain(t, []byte(`{"version":1,"source":"dGVzdA==","steps":[{"plugin":"base64"}]}`))
	var stdout, stderr bytes.Buffer
	code := runChainWithArgs([]string{chainPath}, strings.NewReader("ignored"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "dGVzdA==" {
		t.Fatalf("stdout = %q, want base64 result", got)
	}
}

func TestRunChainWithStdinOverrideAndNewline(t *testing.T) {
	chainPath := writeTestChain(t, []byte(`{"version":1,"steps":[{"plugin":"base64"}]}`))
	var stdout, stderr bytes.Buffer
	code := runChainWithArgs([]string{"-stdin", "-N", chainPath}, strings.NewReader("test"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "dGVzdA==\n" {
		t.Fatalf("stdout = %q, want base64 result with newline", got)
	}
}

func TestRunChainReportsStepError(t *testing.T) {
	chainPath := writeTestChain(t, []byte(`{"version":1,"source":"JSUlJQ==","steps":[{"plugin":"base64","unprocess":true}]}`))
	var stdout, stderr bytes.Buffer
	code := runChainWithArgs([]string{chainPath}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "step 1 (base64)") {
		t.Fatalf("stderr = %q, want step error", stderr.String())
	}
}

func TestRunChainMissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runChainWithArgs(nil, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing chain file") {
		t.Fatalf("stderr = %q, want missing chain file", stderr.String())
	}
}

func TestFirstChainError(t *testing.T) {
	p := pipeline.New()
	p.SetSource([]byte("%%%%"))
	p.AddStep("base64", true)
	i, err := firstChainError(p)
	if err == nil {
		t.Fatal("expected chain error")
	}
	if i != 0 {
		t.Fatalf("error step = %d, want 0", i)
	}
	if !strings.Contains(err.Error(), "base64") {
		t.Fatalf("chain error = %q", err)
	}
}

func writeTestChain(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "chain.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
