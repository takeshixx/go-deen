package hashs

import (
	"crypto/md5"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginMD4 creates a plugin
func NewPluginMD4() *types.DeenPlugin {
	return hashPlugin("md4",
		"MD4 Message-Digest Algorithm is a cryptographic hash function\nwith a digest length of 128 bits.",
		nil, md4.New)
}

// NewPluginMD5 creates a plugin
func NewPluginMD5() *types.DeenPlugin {
	return hashPlugin("md5",
		"MD5 Message-Digest Algorithm is a cryptographic hash function\nwith a digest length of 128 bits.",
		nil, md5.New)
}

// NewPluginRIPEMD160 creates a plugin
func NewPluginRIPEMD160() *types.DeenPlugin {
	return hashPlugin("ripemd160",
		"RIPEMD (RIPE Message Digest) is a family of cryptographic hash\nfunctions developed in 1992.",
		[]string{"md160"}, ripemd160.New)
}
