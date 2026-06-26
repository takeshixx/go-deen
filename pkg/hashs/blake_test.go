package hashs

import "testing"

var blakeTestData = []byte("deenblaketest")
var blakeTestKey = "testkey123"
var blakeTestKey32 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA__"

func TestPluginBLAKE2s(t *testing.T) {
	assertHash(t, NewPluginBLAKE2s(), blakeTestData, "4088080c149a6165b9a086ef4aaeb13df5fc7ffb83d5731ed9692320b5634c50")
	assertHash(t, NewPluginBLAKE2s(), blakeTestData, "183f16fed32775091e9f63171cbdca57e6c0642dffa08c7a25ea58aa7b50f2d5", "-key", blakeTestKey)
	assertHash(t, NewPluginBLAKE2s(), blakeTestData, "b789903bd37697727692abb9f0494bad", "-len", "16", "-key", blakeTestKey)
}

func TestPluginBLAKE2b(t *testing.T) {
	assertHash(t, NewPluginBLAKE2b(), blakeTestData, "e3e8bca1c407f1ce36642d64c334bbc572f7ad06e00425d2abc567e094e9e82862b3d8f200647273ec4f1d36cc5b7371b6a4cf7ea6725529ce71ea9c68eeb66c")
	assertHash(t, NewPluginBLAKE2b(), blakeTestData, "4e8c474aa515d314feb9cb0893e2bddaad49f007fbd1f0538776f2c11d9c9d04732b70a024642400b14707276928c94429b109424245156e438503aa312036d9", "-key", blakeTestKey, "-len", "64")
}

func TestPluginBLAKE2x(t *testing.T) {
	assertHash(t, NewPluginBLAKE2x(), blakeTestData, "0924e4d71784282e91639a595475a0290a9c2caee4a03978199b4d2f7bcf8d83")
	assertHash(t, NewPluginBLAKE2x(), blakeTestData, "94e30c637ecf91f9873924d74e667e56099045b8e3cffaa3ad9d415b163af0ad6c0c1da67732b0d2497f152f197635d2ac76cead6c7a48fcb8c0b11ca3a726f0", "-key", blakeTestKey, "-len", "64")
}

func TestPluginBLAKE3(t *testing.T) {
	assertHash(t, NewPluginBLAKE3(), blakeTestData, "e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b")
	assertHash(t, NewPluginBLAKE3(), blakeTestData, "e60cd8431ebc2c74d793a4e7256344fb5b4050311f3203a3f62eacdc608bd78b9d34e700bc948d24b00be997822acdad00757bd4364cbd5d994531fa492cafa3", "-length", "64")
	assertHash(t, NewPluginBLAKE3(), blakeTestData, "7128ffee8eb9e5eca0bb", "-derive-key", blakeTestKey)
	assertHash(t, NewPluginBLAKE3(), blakeTestData, "654833e7182b2fe8c8ec", "-derive-key", blakeTestKey, "-context", "test context 123")
	assertHash(t, NewPluginBLAKE3(), blakeTestData, "cfa55ce67ffc5b8c45bd9d5fa62947e7246783166b1c649fd8f74771919f90e7", "-key", blakeTestKey32)

	if _, err := tryHash(NewPluginBLAKE3(), blakeTestData, "-key", blakeTestKey); err == nil {
		t.Error("BLAKE3 did not return an error with a short key")
	}
}
