package hashs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginMD4 creates a plugin
func NewPluginMD4() (p types.DeenPlugin) {
	p.Name = "md4"
	p.Aliases = []string{}
	p.Type = "hash"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			hasher := md4.New()
			_, err := io.Copy(hasher, task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into encoder in MD4 failed")
			}
			hashSum := hasher.Sum(nil)
			encodedBuf := make([]byte, hex.EncodedLen(len(hashSum[:])))
			_ = hex.Encode(encodedBuf, hashSum[:])
			outBuf := bytes.NewBuffer(encodedBuf)
			_, err = io.Copy(task.PipeWriter, outBuf)
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter in MD4 failed")
			}
		}()
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
