package hashs

import "testing"

func TestNewPluginSHA224(t *testing.T) {
	assertHash(t, NewPluginSHA224(), shaTestData, "b1dc6360c8531f1b465a484dbea1c5cec454ba3ca29c6eb4cd5ae406")
}

func TestNewPluginSHA256(t *testing.T) {
	assertHash(t, NewPluginSHA256(), shaTestData, "4ada38e80388198b04707df9c7bc6f2d2c3614fc26e7bbf53494008204d80519")
}

func TestNewPluginSHA384(t *testing.T) {
	assertHash(t, NewPluginSHA384(), shaTestData, "e78f30f30f042989efccc643fb310aef66f7602587d92be0657fcb080ab6bf9dea1df62389e70882812dc446587ea7b8")
}

func TestNewPluginSHA512(t *testing.T) {
	assertHash(t, NewPluginSHA512(), shaTestData, "695359c0ba4b7cb76c5287e14c5f2d5284bfa0b5df81dbb2abfab080221019ed9de0a3f3d4307772cf8bc40c16930d4f1b2a0bd0d81e8a9bed2290f588d2d90b")
}
