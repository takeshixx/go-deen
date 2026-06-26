package hashs

import "testing"

func TestPluginHMAC(t *testing.T) {
	assertHash(t, NewPluginHMAC(), shaTestData,
		"e42ba5b146de9bc40d9714a439b2f3ac07c88c9bd4c25865619e7abfc61b0e1d", "-key", "secret")
	assertHash(t, NewPluginHMAC(), shaTestData,
		"dc60e00d2535bee52a414b80e79c5b73c6fcb8fd", "-alg", "sha1", "-key", "secret")
}

func TestPluginHMACInvalidAlg(t *testing.T) {
	if _, err := tryHash(NewPluginHMAC(), shaTestData, "-alg", "bogus", "-key", "k"); err == nil {
		t.Error("expected an error for an unsupported algorithm")
	}
}
