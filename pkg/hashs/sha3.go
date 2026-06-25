package hashs

import (
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/sha3"
)

const sha3Description = "SHA3 is the latest member of the Secure Hash Algorithm family of\nstandards, released by NIST."

// NewPluginSHA3224 creates a plugin
func NewPluginSHA3224() *types.DeenPlugin {
	return hashPlugin("sha3-224", sha3Description, nil, sha3.New224)
}

// NewPluginSHA3256 creates a plugin
func NewPluginSHA3256() *types.DeenPlugin {
	return hashPlugin("sha3-256", sha3Description, nil, sha3.New256)
}

// NewPluginSHA3384 creates a plugin
func NewPluginSHA3384() *types.DeenPlugin {
	return hashPlugin("sha3-384", sha3Description, nil, sha3.New384)
}

// NewPluginSHA3512 creates a plugin
func NewPluginSHA3512() *types.DeenPlugin {
	return hashPlugin("sha3-512", sha3Description, nil, sha3.New512)
}
