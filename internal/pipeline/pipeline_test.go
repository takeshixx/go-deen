package pipeline

import (
	"bytes"
	"testing"
)

func TestChainEncode(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false) // -> dGVzdA==
	p.AddStep("hex", false)    // -> hex of dGVzdA==

	if got := string(p.Output(0)); got != "dGVzdA==" {
		t.Errorf("step 0: got %q", got)
	}
	// hex of "dGVzdA==" is 6447567a64413d3d
	if got := string(p.Output(1)); got != "6447567a64413d3d" {
		t.Errorf("step 1: got %q", got)
	}
}

func TestChainRecomputesOnSourceChange(t *testing.T) {
	p := New()
	p.SetSource([]byte("a"))
	p.AddStep("base64", false)
	first := string(p.Result())
	p.SetSource([]byte("b"))
	if string(p.Result()) == first {
		t.Error("result did not change after source edit")
	}
	if got := string(p.Result()); got != "Yg==" {
		t.Errorf("got %q, want Yg==", got)
	}
}

func TestDecodeDirection(t *testing.T) {
	p := New()
	p.SetSource([]byte("dGVzdA=="))
	p.AddStep("base64", true) // decode
	if got := string(p.Result()); got != "test" {
		t.Errorf("decode: got %q", got)
	}
}

func TestOneWayDecodeError(t *testing.T) {
	p := New()
	p.SetSource([]byte("x"))
	p.AddStep("sha256", true) // hashes can't decode
	if p.Err(0) == nil {
		t.Error("expected an error decoding with a one-way plugin")
	}
}

func TestOptionsApplied(t *testing.T) {
	p := New()
	p.SetSource([]byte("<<<>>>"))
	i := p.AddStep("base64", false)
	std := string(p.Output(i))
	p.SetOption(i, "url", "true")
	url := string(p.Output(i))
	if std == url {
		t.Error("setting -url did not change base64 output")
	}
}

func TestEditOutputRecomputesDownstream(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false) // step 0 -> dGVzdA==
	p.AddStep("base64", true)  // step 1 decodes -> test

	// Override step 0 with a different base64 value; step 1 must follow.
	p.EditOutput(0, []byte("YQ==")) // base64("a")
	if got := string(p.Output(0)); got != "YQ==" {
		t.Errorf("override not applied: %q", got)
	}
	if got := string(p.Output(1)); got != "a" {
		t.Errorf("downstream did not recompute: %q", got)
	}

	// Editing the source clears the override.
	p.SetSource([]byte("test"))
	if got := string(p.Output(0)); got != "dGVzdA==" {
		t.Errorf("override not cleared on source edit: %q", got)
	}
}

func TestRemoveStep(t *testing.T) {
	p := New()
	p.SetSource([]byte("test"))
	p.AddStep("base64", false)
	p.AddStep("hex", false)
	p.RemoveStep(1)
	if p.Len() != 1 {
		t.Fatalf("expected 1 step, got %d", p.Len())
	}
	if got := string(p.Result()); got != "dGVzdA==" {
		t.Errorf("after remove: got %q", got)
	}
}

func TestBinarySafeChain(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}
	p := New()
	p.SetSource(src)
	p.AddStep("gzip", false)
	p.AddStep("gzip", true)
	if !bytes.Equal(p.Result(), src) {
		t.Error("binary round-trip through the pipeline failed")
	}
}

func TestPluginOptions(t *testing.T) {
	opts := PluginOptions("base64")
	if len(opts) == 0 {
		t.Fatal("base64 should report options")
	}
	var foundURL bool
	for _, o := range opts {
		if o.Name == "url" {
			foundURL = true
			if !o.IsBool {
				t.Error("-url should be a bool option")
			}
		}
	}
	if !foundURL {
		t.Error("base64 options missing -url")
	}
	if PluginOptions("sha256") != nil {
		t.Error("sha256 should have no options")
	}
}
