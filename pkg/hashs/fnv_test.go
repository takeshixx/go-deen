package hashs

import "testing"

func TestPluginFNV(t *testing.T) {
	assertHash(t, NewPluginFNV32(), shaTestData, "ca4d79e9")
	assertHash(t, NewPluginFNV32a(), shaTestData, "dad60189")
	assertHash(t, NewPluginFNV64(), shaTestData, "2fdadc29513dbda9")
	assertHash(t, NewPluginFNV64a(), shaTestData, "7ffbb115d7514a49")
	assertHash(t, NewPluginFNV128(), shaTestData, "fefc5cc7c05353aff890fb90a8c5c029")
	assertHash(t, NewPluginFNV128a(), shaTestData, "003a7c4acf2ea054f8d85f27cf3037b9")
}
