package hashs

import (
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginAdler32 creates a plugin
func NewPluginAdler32() *types.DeenPlugin {
	return hashPlugin("adler32",
		"Adler-32 checksum as defined in RFC 1950.",
		nil, func() hash.Hash { return adler32.New() })
}

// NewPluginCRC32 creates a plugin computing the IEEE CRC-32 checksum.
func NewPluginCRC32() *types.DeenPlugin {
	return hashPlugin("crc32",
		"CRC-32 checksum using the IEEE polynomial (used by zlib, gzip, PNG).",
		nil, func() hash.Hash { return crc32.NewIEEE() })
}

// NewPluginCRC32C creates a plugin computing the Castagnoli CRC-32 checksum.
func NewPluginCRC32C() *types.DeenPlugin {
	tab := crc32.MakeTable(crc32.Castagnoli)
	return hashPlugin("crc32c",
		"CRC-32 checksum using the Castagnoli polynomial.",
		nil, func() hash.Hash { return crc32.New(tab) })
}

// NewPluginCRC32Koopman creates a plugin computing the Koopman CRC-32 checksum.
func NewPluginCRC32Koopman() *types.DeenPlugin {
	tab := crc32.MakeTable(crc32.Koopman)
	return hashPlugin("crc32k",
		"CRC-32 checksum using the Koopman polynomial.",
		nil, func() hash.Hash { return crc32.New(tab) })
}

// NewPluginCRC64ISO creates a plugin computing the ISO CRC-64 checksum.
func NewPluginCRC64ISO() *types.DeenPlugin {
	tab := crc64.MakeTable(crc64.ISO)
	return hashPlugin("crc64",
		"CRC-64 checksum using the ISO polynomial.",
		[]string{"crc64-iso"}, func() hash.Hash { return crc64.New(tab) })
}

// NewPluginCRC64ECMA creates a plugin computing the ECMA CRC-64 checksum.
func NewPluginCRC64ECMA() *types.DeenPlugin {
	tab := crc64.MakeTable(crc64.ECMA)
	return hashPlugin("crc64-ecma",
		"CRC-64 checksum using the ECMA polynomial.",
		nil, func() hash.Hash { return crc64.New(tab) })
}
