package hashs

import "testing"

var mdTestData = []byte("deenmdtest")

func TestNewPluginMD4(t *testing.T) {
	assertHash(t, NewPluginMD4(), mdTestData, "a635f9247276ff156bbbb3752db8a2b1")
}

func TestNewPluginMD5(t *testing.T) {
	assertHash(t, NewPluginMD5(), mdTestData, "204778cb29e3c5ffa8037312a6ea2a56")
}

func TestNewPluginRIPEMD160(t *testing.T) {
	assertHash(t, NewPluginRIPEMD160(), mdTestData, "8bcf2832074848df4727aeb275e4905f2213814e")
}
