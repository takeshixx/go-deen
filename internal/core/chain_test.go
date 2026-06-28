package core

import (
	"io"
	"strings"
	"testing"

	"github.com/takeshixx/deen/internal/pipeline"
)

func TestSelectChainInputUsesArgs(t *testing.T) {
	r, cleanup, err := selectChainInput("", false, []string{"hello", "world"})
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
