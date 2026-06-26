package compressions

import "testing"

func TestPluginZstd(t *testing.T) {
	assertRoundTrip(t, NewPluginZstd())
	assertDecompressError(t, NewPluginZstd())
}
