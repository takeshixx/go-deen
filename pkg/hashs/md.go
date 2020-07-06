package hashs

import (
	"crypto/md5"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"

	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginMD4 creates a plugin
func NewPluginMD4() (p types.DeenPlugin) {
	p.Name = "md4"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := md4.New()
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

// NewPluginMD5 creates a plugin
func NewPluginMD5() (p types.DeenPlugin) {
	p.Name = "md5"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := md5.New()
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

// NewPluginRIPEMD160 creates a plugin
func NewPluginRIPEMD160() (p types.DeenPlugin) {
	p.Name = "ripemd160"
	p.Aliases = []string{"md160"}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var err error
		hasher := ripemd160.New()
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
