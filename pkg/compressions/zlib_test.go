package compressions

import "testing"

func TestPluginZlib(t *testing.T) {
	assertRoundTrip(t, NewPluginZlib())
	assertRoundTrip(t, NewPluginZlib(), "-level", "9")
	assertDecompressError(t, NewPluginZlib())
}
