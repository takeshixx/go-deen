package compressions

import "testing"

func TestPluginGzip(t *testing.T) {
	assertRoundTrip(t, NewPluginGzip())
	assertRoundTrip(t, NewPluginGzip(), "-level", "9")
	assertDecompressError(t, NewPluginGzip())
}
