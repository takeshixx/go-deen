package hashs

import (
	"crypto/sha1"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSHA1 creates a plugin
func NewPluginSHA1() *types.DeenPlugin {
	return hashPlugin("sha1",
		"SHA1 is a cryptographic hash function which takes an input and\nproduces a 160-bit (20-byte) hash value known as a message digest.",
		nil, sha1.New)
}
