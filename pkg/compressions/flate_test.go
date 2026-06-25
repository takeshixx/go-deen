package compressions

import "testing"

func TestPluginFlate(t *testing.T) {
	assertRoundTrip(t, NewPluginFlate())
	assertRoundTrip(t, NewPluginFlate(), "-level", "9")
}
