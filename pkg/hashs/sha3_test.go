package hashs

import "testing"

func TestNewPluginSHA3224(t *testing.T) {
	assertHash(t, NewPluginSHA3224(), shaTestData, "4c6f98b0df46532ed5f0b30de780362d3bdadb08ae6ff9f0a0dc468b")
}

func TestNewPluginSHA3256(t *testing.T) {
	assertHash(t, NewPluginSHA3256(), shaTestData, "281ead2ece78635fa0ec9cfe26b0df3342ccceffd630890597cd3d34bec9ad58")
}

func TestNewPluginSHA3384(t *testing.T) {
	assertHash(t, NewPluginSHA3384(), shaTestData, "0f7aa1bd9ccaf36dbdbe965f0a98394ca3aad8a9ea6a6d13fb6783f45b73f5c73c333f2bf063e070ecc8fd47602b70a5")
}

func TestNewPluginSHA3512(t *testing.T) {
	assertHash(t, NewPluginSHA3512(), shaTestData, "454c165b1f13db38ce7fa7dbad1d4c9f26b7cc085832b32c6fb1c965eb64894ac9133f9ad7691cc1da7ff95bbbd2259df898ac6686b692f066a62ec996b9b4ec")
}
