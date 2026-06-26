package hashs

import (
	"hash"
	"hash/fnv"

	"github.com/takeshixx/deen/pkg/types"
)

const fnvDescription = "Fowler-Noll-Vo (FNV) is a non-cryptographic hash function."

// NewPluginFNV32 creates a plugin computing the 32-bit FNV-1 hash.
func NewPluginFNV32() *types.DeenPlugin {
	return hashPlugin("fnv32", fnvDescription, nil, func() hash.Hash { return fnv.New32() })
}

// NewPluginFNV32a creates a plugin computing the 32-bit FNV-1a hash.
func NewPluginFNV32a() *types.DeenPlugin {
	return hashPlugin("fnv32a", fnvDescription, nil, func() hash.Hash { return fnv.New32a() })
}

// NewPluginFNV64 creates a plugin computing the 64-bit FNV-1 hash.
func NewPluginFNV64() *types.DeenPlugin {
	return hashPlugin("fnv64", fnvDescription, nil, func() hash.Hash { return fnv.New64() })
}

// NewPluginFNV64a creates a plugin computing the 64-bit FNV-1a hash.
func NewPluginFNV64a() *types.DeenPlugin {
	return hashPlugin("fnv64a", fnvDescription, nil, func() hash.Hash { return fnv.New64a() })
}

// NewPluginFNV128 creates a plugin computing the 128-bit FNV-1 hash.
func NewPluginFNV128() *types.DeenPlugin {
	return hashPlugin("fnv128", fnvDescription, nil, fnv.New128)
}

// NewPluginFNV128a creates a plugin computing the 128-bit FNV-1a hash.
func NewPluginFNV128a() *types.DeenPlugin {
	return hashPlugin("fnv128a", fnvDescription, nil, fnv.New128a)
}
