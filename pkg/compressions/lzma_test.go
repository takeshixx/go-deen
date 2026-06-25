package compressions

import "testing"

func TestPluginLZMA(t *testing.T) {
	assertRoundTrip(t, NewPluginLZMA())
}

func TestPluginLZMA2(t *testing.T) {
	assertRoundTrip(t, NewPluginLZMA2())
}
