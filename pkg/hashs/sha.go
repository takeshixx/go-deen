package hashs

import (
	"crypto/sha1"
	"encoding/hex"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSHA1 creates a plugin
func NewPluginSHA1() (p types.DeenPlugin) {
	p.Name = "sha1"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha1.New()
		if _, err := io.Copy(hasher, reader); err != nil {
			return *new([]byte), err
		}
		hashSum := hasher.Sum(nil)
		outBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
		_ = hex.Encode(outBuf, hashSum[:])
		return outBuf, err
	}
	return
}
