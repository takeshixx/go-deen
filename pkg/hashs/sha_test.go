package hashs

import "testing"

var shaTestData = []byte("deenshatest")

func TestNewPluginSHA1(t *testing.T) {
	assertHash(t, NewPluginSHA1(), shaTestData, "c324da7d32853ffaeb6577f624753c7f0f2842c0")
}
