package hashs

import "testing"

func TestPluginChecksums(t *testing.T) {
	assertHash(t, NewPluginAdler32(), shaTestData, "1b190499")
	assertHash(t, NewPluginCRC32(), shaTestData, "0b9f556c")
	assertHash(t, NewPluginCRC32C(), shaTestData, "5ae313e8")
	assertHash(t, NewPluginCRC32Koopman(), shaTestData, "fc410794")
	assertHash(t, NewPluginCRC64ISO(), shaTestData, "f3b0bf6f64183c7d")
	assertHash(t, NewPluginCRC64ECMA(), shaTestData, "7759fbe15fe30f39")
}
