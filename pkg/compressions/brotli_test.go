package compressions

import "testing"

func TestPluginBrotli(t *testing.T) {
	assertRoundTrip(t, NewPluginBrotli())
	assertRoundTrip(t, NewPluginBrotli(), "-level", "11")
}

func TestPluginBrotliInvalidLevel(t *testing.T) {
	p := NewPluginBrotli()
	if _, err := transform(p.Process, p.RegisterFlags, compTestData, "-level", "99"); err == nil {
		t.Error("expected an error for an out-of-range level")
	}
}
