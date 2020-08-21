package hashs

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"io"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginSHA224 creates a plugin
func NewPluginSHA224() (p types.DeenPlugin) {
	p.Name = "sha224"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha256.New224()
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

// NewPluginSHA256 creates a plugin
func NewPluginSHA256() (p types.DeenPlugin) {
	p.Name = "sha256"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha256.New()
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

// NewPluginSHA384 creates a plugin
func NewPluginSHA384() (p types.DeenPlugin) {
	p.Name = "sha384"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha512.New384()
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

// NewPluginSHA512 creates a plugin
func NewPluginSHA512() (p types.DeenPlugin) {
	p.Name = "sha512"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := sha512.New()
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
