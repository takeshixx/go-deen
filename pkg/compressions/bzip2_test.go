package compressions

import "testing"

func TestPluginBzip2(t *testing.T) {
	assertRoundTrip(t, NewPluginBzip2())
	assertRoundTrip(t, NewPluginBzip2(), "-level", "9")
	assertDecompressError(t, NewPluginBzip2())
}

func TestPluginBzip2InvalidLevel(t *testing.T) {
	p := NewPluginBzip2()
	if _, err := transform(p.Process, p.RegisterFlags, compTestData, "-level", "99"); err == nil {
		t.Error("expected an error for an out-of-range level")
	}
}
