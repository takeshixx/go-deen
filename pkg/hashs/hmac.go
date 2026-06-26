package hashs

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"sort"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/sha3"
)

// hmacHashes maps the -alg flag value to the underlying hash constructor.
var hmacHashes = map[string]func() hash.Hash{
	"md5":      md5.New,
	"sha1":     sha1.New,
	"sha224":   sha256.New224,
	"sha256":   sha256.New,
	"sha384":   sha512.New384,
	"sha512":   sha512.New,
	"sha3-256": sha3.New256,
	"sha3-512": sha3.New512,
}

func hmacAlgNames() string {
	names := make([]string, 0, len(hmacHashes))
	for n := range hmacHashes {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// NewPluginHMAC creates a keyed-hash message authentication code plugin.
func NewPluginHMAC() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "hmac"
	p.Category = "hashs"
	p.Description = "Keyed-Hash Message Authentication Code (HMAC, RFC 2104)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("alg", "sha256", "hash algorithm ("+hmacAlgNames()+")")
		flags.String("key", "", "secret key")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		alg := helpers.StringFlag(flags, "alg")
		newHash, ok := hmacHashes[alg]
		if !ok {
			return fmt.Errorf("unsupported algorithm %q (supported: %s)", alg, hmacAlgNames())
		}
		mac := hmac.New(newHash, []byte(helpers.StringFlag(flags, "key")))
		if _, err := io.Copy(mac, r); err != nil {
			return err
		}
		_, err := io.WriteString(w, hex.EncodeToString(mac.Sum(nil)))
		return err
	}
	return p
}
