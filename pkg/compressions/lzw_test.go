package compressions

import "testing"

func TestPluginLzw(t *testing.T) {
	assertRoundTrip(t, NewPluginLzw())
	assertRoundTrip(t, NewPluginLzw(), "-order", "1")
}
