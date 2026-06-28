package misc

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
	"golang.org/x/crypto/chacha20poly1305"
)

func bytesFlag(flags *flag.FlagSet, name string) ([]byte, error) {
	value := strings.TrimSpace(helpers.StringFlag(flags, name))
	if value == "" {
		return nil, fmt.Errorf("missing -%s", name)
	}
	return decodeHexOrBase64(value)
}

func bytesAliasFlag(flags *flag.FlagSet, primary, alias string) ([]byte, error) {
	primaryValue := strings.TrimSpace(helpers.StringFlag(flags, primary))
	aliasValue := strings.TrimSpace(helpers.StringFlag(flags, alias))
	if primaryValue == "" && aliasValue == "" {
		return nil, fmt.Errorf("missing -%s or -%s", primary, alias)
	}
	if primaryValue == "" {
		return decodeHexOrBase64(aliasValue)
	}
	primaryBytes, err := decodeHexOrBase64(primaryValue)
	if err != nil {
		return nil, fmt.Errorf("invalid -%s: %w", primary, err)
	}
	if aliasValue == "" {
		return primaryBytes, nil
	}
	aliasBytes, err := decodeHexOrBase64(aliasValue)
	if err != nil {
		return nil, fmt.Errorf("invalid -%s: %w", alias, err)
	}
	if !bytes.Equal(primaryBytes, aliasBytes) {
		return nil, fmt.Errorf("conflicting -%s and -%s values", primary, alias)
	}
	return primaryBytes, nil
}

func decodeHexOrBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty value")
	}
	if b, err := hex.DecodeString(value); err == nil {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(value); err == nil {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(value); err == nil {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(value); err == nil {
		return b, nil
	}
	return base64.RawURLEncoding.DecodeString(value)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	n := blockSize - len(data)%blockSize
	return append(append([]byte(nil), data...), bytes.Repeat([]byte{byte(n)}, n)...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padded data length")
	}
	n := int(data[len(data)-1])
	if n == 0 || n > blockSize || n > len(data) {
		return nil, fmt.Errorf("invalid padding")
	}
	for _, b := range data[len(data)-n:] {
		if int(b) != n {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-n], nil
}

// NewPluginAES creates an AES encryption/decryption plugin for GCM, CBC and CTR.
func NewPluginAES() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "aes"
	p.Category = "misc"
	p.Description = "Encrypt with AES-GCM, AES-CBC or AES-CTR; decode mode decrypts."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("mode", "gcm", "AES mode: gcm, cbc or ctr")
		flags.String("key", "", "hex/base64 AES key: 16, 24 or 32 bytes")
		flags.String("nonce", "", "hex/base64 GCM nonce")
		flags.String("iv", "", "hex/base64 CBC/CTR IV")
		flags.String("aad", "", "additional authenticated data for GCM")
		flags.Int("tag-len", 16, "GCM authentication tag length in bytes: 12-16")
		flags.Bool("skip-aead-verify", false, "decrypt GCM ciphertext without verifying the authentication tag")
		flags.String("padding", "pkcs7", "CBC padding: pkcs7 or none")
	}
	p.Process = aesTransform(false)
	p.Unprocess = aesTransform(true)
	return p
}

func aesTransform(decrypt bool) types.TransformFunc {
	return func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		key, err := bytesFlag(flags, "key")
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		switch strings.ToLower(helpers.StringFlag(flags, "mode")) {
		case "", "gcm":
			nonce, err := bytesAliasFlag(flags, "nonce", "iv")
			if err != nil {
				return err
			}
			tagLen := helpers.IntFlag(flags, "tag-len", 16)
			if tagLen < 12 || tagLen > 16 {
				return fmt.Errorf("GCM tag length must be between 12 and 16 bytes")
			}
			aead, err := cipher.NewGCMWithTagSize(block, tagLen)
			if err != nil {
				return err
			}
			if len(nonce) != aead.NonceSize() {
				return fmt.Errorf("nonce must be %d bytes", aead.NonceSize())
			}
			if decrypt {
				ciphertext := input
				input, err = aead.Open(nil, nonce, ciphertext, []byte(helpers.StringFlag(flags, "aad")))
				if err != nil && helpers.IsBoolFlag(flags, "skip-aead-verify") {
					plain, unsafeErr := unsafeGCMPlaintext(block, nonce, ciphertext, tagLen)
					if unsafeErr != nil {
						return unsafeErr
					}
					if _, writeErr := w.Write(plain); writeErr != nil {
						return writeErr
					}
					return fmt.Errorf("%w; displayed unauthenticated plaintext because -skip-aead-verify is set", err)
				}
			} else {
				input = aead.Seal(nil, nonce, input, []byte(helpers.StringFlag(flags, "aad")))
			}
			if err != nil {
				return err
			}
			_, err = w.Write(input)
			return err
		case "cbc":
			iv, err := bytesAliasFlag(flags, "iv", "nonce")
			if err != nil {
				return err
			}
			if len(iv) != block.BlockSize() {
				return fmt.Errorf("iv must be %d bytes", block.BlockSize())
			}
			padding := strings.ToLower(strings.TrimSpace(helpers.StringFlag(flags, "padding")))
			switch padding {
			case "", "pkcs", "pkcs5", "pkcs7":
				padding = "pkcs7"
			case "none", "no":
				padding = "none"
			default:
				return fmt.Errorf("unsupported CBC padding")
			}
			if decrypt {
				if len(input)%block.BlockSize() != 0 {
					return fmt.Errorf("ciphertext is not a multiple of block size")
				}
				out := append([]byte(nil), input...)
				cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, out)
				if padding == "pkcs7" {
					out, err = pkcs7Unpad(out, block.BlockSize())
					if err != nil {
						return err
					}
				}
				_, err = w.Write(out)
				return err
			}
			out := append([]byte(nil), input...)
			if padding == "pkcs7" {
				out = pkcs7Pad(input, block.BlockSize())
			} else if len(out)%block.BlockSize() != 0 {
				return fmt.Errorf("plaintext is not a multiple of block size")
			}
			cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, out)
			_, err = w.Write(out)
			return err
		case "ctr":
			iv, err := bytesAliasFlag(flags, "iv", "nonce")
			if err != nil {
				return err
			}
			if len(iv) != block.BlockSize() {
				return fmt.Errorf("iv must be %d bytes", block.BlockSize())
			}
			out := make([]byte, len(input))
			cipher.NewCTR(block, iv).XORKeyStream(out, input)
			_, err = w.Write(out)
			return err
		default:
			return fmt.Errorf("unsupported AES mode")
		}
	}
}

func unsafeGCMPlaintext(block cipher.Block, nonce, ciphertext []byte, tagLen int) ([]byte, error) {
	if len(ciphertext) < tagLen {
		return nil, fmt.Errorf("ciphertext is shorter than GCM tag length")
	}
	ciphertext = ciphertext[:len(ciphertext)-tagLen]
	counter := make([]byte, block.BlockSize())
	copy(counter, nonce)
	binary.BigEndian.PutUint32(counter[len(counter)-4:], 2)
	plain := make([]byte, len(ciphertext))
	cipher.NewCTR(block, counter).XORKeyStream(plain, ciphertext)
	return plain, nil
}

// NewPluginChaCha20Poly1305 creates a ChaCha20-Poly1305 AEAD plugin.
func NewPluginChaCha20Poly1305() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "chacha20poly1305"
	p.Aliases = []string{"chacha"}
	p.Category = "misc"
	p.Description = "Encrypt with ChaCha20-Poly1305; decode mode decrypts."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("key", "", "hex/base64 32-byte key")
		flags.String("nonce", "", "hex/base64 12-byte nonce")
		flags.String("aad", "", "additional authenticated data")
	}
	p.Process = chachaTransform(false)
	p.Unprocess = chachaTransform(true)
	return p
}

func chachaTransform(decrypt bool) types.TransformFunc {
	return func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		key, err := bytesFlag(flags, "key")
		if err != nil {
			return err
		}
		nonce, err := bytesFlag(flags, "nonce")
		if err != nil {
			return err
		}
		aead, err := chacha20poly1305.New(key)
		if err != nil {
			return err
		}
		if len(nonce) != aead.NonceSize() {
			return fmt.Errorf("nonce must be %d bytes", aead.NonceSize())
		}
		if decrypt {
			input, err = aead.Open(nil, nonce, input, []byte(helpers.StringFlag(flags, "aad")))
		} else {
			input = aead.Seal(nil, nonce, input, []byte(helpers.StringFlag(flags, "aad")))
		}
		if err != nil {
			return err
		}
		_, err = w.Write(input)
		return err
	}
}

type ecdsaSignature struct {
	R, S *big.Int
}

// NewPluginSign creates a sign/verify plugin for Ed25519, RSA-PSS and ECDSA.
func NewPluginSign() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "sign"
	p.Aliases = []string{"verify"}
	p.Category = "misc"
	p.Description = "Sign input; decode mode verifies signatures. Supports ed25519, rsa-pss and ecdsa."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("alg", "ed25519", "algorithm: ed25519, rsa-pss or ecdsa")
		flags.String("key", "", "private key: PEM path or hex/base64 Ed25519 private key/seed")
		flags.String("pub", "", "public key: PEM path or hex/base64 Ed25519 public key")
		flags.String("sig", "", "signature as hex/base64 or file path when verifying")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		sig, err := signData(strings.ToLower(helpers.StringFlag(flags, "alg")), helpers.StringFlag(flags, "key"), input)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, hex.EncodeToString(sig))
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		sig, err := readHexOrFile(helpers.StringFlag(flags, "sig"))
		if err != nil {
			return err
		}
		if err := verifyData(strings.ToLower(helpers.StringFlag(flags, "alg")), helpers.StringFlag(flags, "pub"), input, sig); err != nil {
			return err
		}
		_, err = io.WriteString(w, "valid")
		return err
	}
	return p
}

func readHexOrFile(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("missing value")
	}
	if b, err := hex.DecodeString(value); err == nil {
		return b, nil
	}
	if b, err := decodeHexOrBase64(value); err == nil {
		return b, nil
	}
	return os.ReadFile(value)
}

func readPEMOrHex(value string) ([]byte, error) {
	if b, err := os.ReadFile(value); err == nil {
		return b, nil
	}
	if b, err := hex.DecodeString(strings.TrimSpace(value)); err == nil {
		return b, nil
	}
	if b, err := decodeHexOrBase64(value); err == nil {
		return b, nil
	}
	return []byte(value), nil
}

func parseSignPrivateKey(value string) (interface{}, error) {
	data, err := readPEMOrHex(value)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(data); block != nil {
		if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			return key, nil
		}
		if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			return key, nil
		}
		if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
			return key, nil
		}
		return nil, fmt.Errorf("unsupported private key")
	}
	if len(data) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(data), nil
	}
	if len(data) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(data), nil
	}
	return nil, fmt.Errorf("unsupported private key")
}

func parsePublicKey(value string) (interface{}, error) {
	data, err := readPEMOrHex(value)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(data); block != nil {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	}
	if len(data) == ed25519.PublicKeySize {
		return ed25519.PublicKey(data), nil
	}
	return nil, fmt.Errorf("unsupported public key")
}

func signData(alg, keyValue string, data []byte) ([]byte, error) {
	key, err := parseSignPrivateKey(keyValue)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	switch alg {
	case "", "ed25519":
		k, ok := key.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected Ed25519 private key")
		}
		return ed25519.Sign(k, data), nil
	case "rsa-pss", "rsa":
		k, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA private key")
		}
		return rsa.SignPSS(rand.Reader, k, crypto.SHA256, digest[:], nil)
	case "ecdsa":
		k, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected ECDSA private key")
		}
		r, s, err := ecdsa.Sign(rand.Reader, k, digest[:])
		if err != nil {
			return nil, err
		}
		return asn1.Marshal(ecdsaSignature{r, s})
	default:
		return nil, fmt.Errorf("unsupported algorithm %q", alg)
	}
}

func verifyData(alg, pubValue string, data, sig []byte) error {
	key, err := parsePublicKey(pubValue)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	switch alg {
	case "", "ed25519":
		k, ok := key.(ed25519.PublicKey)
		if !ok {
			return fmt.Errorf("expected Ed25519 public key")
		}
		if !ed25519.Verify(k, data, sig) {
			return fmt.Errorf("invalid signature")
		}
		return nil
	case "rsa-pss", "rsa":
		k, ok := key.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("expected RSA public key")
		}
		return rsa.VerifyPSS(k, crypto.SHA256, digest[:], sig, nil)
	case "ecdsa":
		k, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("expected ECDSA public key")
		}
		var esig ecdsaSignature
		if _, err := asn1.Unmarshal(sig, &esig); err != nil {
			return err
		}
		if !ecdsa.Verify(k, digest[:], esig.R, esig.S) {
			return fmt.Errorf("invalid signature")
		}
		return nil
	default:
		return fmt.Errorf("unsupported algorithm %q", alg)
	}
}
