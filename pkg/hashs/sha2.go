package hashs

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/takeshixx/deen/pkg/types"
)

const sha2Description = "SHA2 is a set of cryptographic hash functions designed by the\nUnited States National Security Agency (NSA)."

// NewPluginSHA224 creates a plugin
func NewPluginSHA224() *types.DeenPlugin {
	return hashPlugin("sha224", sha2Description, nil, sha256.New224)
}

// NewPluginSHA256 creates a plugin
func NewPluginSHA256() *types.DeenPlugin {
	return hashPlugin("sha256", sha2Description, nil, sha256.New)
}

// NewPluginSHA384 creates a plugin
func NewPluginSHA384() *types.DeenPlugin {
	return hashPlugin("sha384", sha2Description, nil, sha512.New384)
}

// NewPluginSHA512 creates a plugin
func NewPluginSHA512() *types.DeenPlugin {
	return hashPlugin("sha512", sha2Description, nil, sha512.New)
}

// NewPluginSHA512_224 creates a plugin
func NewPluginSHA512_224() *types.DeenPlugin {
	return hashPlugin("sha512-224", sha2Description, nil, sha512.New512_224)
}

// NewPluginSHA512_256 creates a plugin
func NewPluginSHA512_256() *types.DeenPlugin {
	return hashPlugin("sha512-256", sha2Description, nil, sha512.New512_256)
}
