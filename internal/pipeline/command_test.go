package pipeline

import "testing"

func TestCommandLine(t *testing.T) {
	p := New()
	p.AddStep("base64", false)
	p.SetOption(0, "url", "true")
	p.AddStep("gzip", true)

	want := "deen base64 -url | deen .gzip"
	if got := p.CommandLine(); got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestCommandLineSkipsDisabledAndEmptySteps(t *testing.T) {
	p := New()
	p.AddStep("base64", false)
	p.AddStep("", false)
	p.AddStep("hex", false)
	p.SetStepDisabled(0, true)

	want := "deen hex"
	if got := p.CommandLine(); got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestCommandLineQuotesOptionValues(t *testing.T) {
	p := New()
	p.AddStep("jq", false)
	p.SetOption(0, "q", `.items[] | select(.name == "a b")`)
	p.SetOption(0, "plain", "true")

	want := `deen jq -plain -q '.items[] | select(.name == "a b")'`
	if got := p.CommandLine(); got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestCommandLineEmptyChain(t *testing.T) {
	if got := New().CommandLine(); got != "" {
		t.Fatalf("empty command = %q", got)
	}
}

func TestShellQuoteSingleQuote(t *testing.T) {
	if got := shellQuote("a'b"); got != `'a'"'"'b'` {
		t.Fatalf("quote = %q", got)
	}
}
