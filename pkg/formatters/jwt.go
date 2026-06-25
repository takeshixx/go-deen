package formatters

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/iancoleman/orderedmap"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var keyManagementAlgs = map[jose.KeyAlgorithm]string{
	jose.ED25519:            "",
	jose.RSA1_5:             "RSA-PKCS1v1.5",
	jose.RSA_OAEP:           "RSA-OAEP-SHA1",
	jose.RSA_OAEP_256:       "RSA-OAEP-SHA256",
	jose.A128KW:             "AES key wrap (128)",
	jose.A192KW:             "AES key wrap (192)",
	jose.A256KW:             "AES key wrap (256)",
	jose.DIRECT:             "Direct encryption",
	jose.ECDH_ES:            "ECDH-ES",
	jose.ECDH_ES_A128KW:     "ECDH-ES + AES key wrap (128)",
	jose.ECDH_ES_A192KW:     "ECDH-ES + AES key wrap (192)",
	jose.ECDH_ES_A256KW:     "ECDH-ES + AES key wrap (256)",
	jose.A128GCMKW:          "AES-GCM key wrap (128)",
	jose.A192GCMKW:          "AES-GCM key wrap (192)",
	jose.A256GCMKW:          "AES-GCM key wrap (256)",
	jose.PBES2_HS256_A128KW: "PBES2 + HMAC-SHA256 + AES key wrap (128)",
	jose.PBES2_HS384_A192KW: "PBES2 + HMAC-SHA384 + AES key wrap (192)",
	jose.PBES2_HS512_A256KW: "PBES2 + HMAC-SHA512 + AES key wrap (256)",
}

var signatureAlgs = map[jose.SignatureAlgorithm]string{
	jose.EdDSA: "",
	jose.HS256: "HMAC using SHA-256",
	jose.HS384: "HMAC using SHA-384",
	jose.HS512: "HMAC using SHA-512",
	jose.RS256: "RSASSA-PKCS-v1.5 using SHA-256",
	jose.RS384: "RSASSA-PKCS-v1.5 using SHA-384",
	jose.RS512: "RSASSA-PKCS-v1.5 using SHA-512",
	jose.ES256: "ECDSA using P-256 and SHA-256",
	jose.ES384: "ECDSA using P-384 and SHA-384",
	jose.ES512: "ECDSA using P-521 and SHA-512",
	jose.PS256: "RSASSA-PSS using SHA256 and MGF1-SHA256",
	jose.PS384: "RSASSA-PSS using SHA384 and MGF1-SHA384",
	jose.PS512: "RSASSA-PSS using SHA512 and MGF1-SHA512",
}

var contentEncryptionAlgs = map[jose.ContentEncryption]string{
	jose.A128CBC_HS256: "AES-CBC + HMAC-SHA256 (128)",
	jose.A192CBC_HS384: "AES-CBC + HMAC-SHA384 (192)",
	jose.A256CBC_HS512: "AES-CBC + HMAC-SHA512 (256)",
	jose.A128GCM:       "AES-GCM (128)",
	jose.A192GCM:       "AES-GCM (192)",
	jose.A256GCM:       "AES-GCM (256)",
}

var compressionAlgs = map[jose.CompressionAlgorithm]string{
	jose.NONE:    "No compression",
	jose.DEFLATE: "DEFLATE (RFC 1951)",
}

// allowedSignatureAlgorithms returns every signature algorithm deen knows
// about. go-jose v4 requires callers of ParseSigned to explicitly allow-list
// the algorithms they expect; deen is a decoding tool that must accept any
// valid token, so we permit all supported algorithms.
func allowedSignatureAlgorithms() []jose.SignatureAlgorithm {
	algs := make([]jose.SignatureAlgorithm, 0, len(signatureAlgs))
	for alg := range signatureAlgs {
		algs = append(algs, alg)
	}
	return algs
}

func listAlgs() string {
	out := "Key management algorithms:\n"
	for k, v := range keyManagementAlgs {
		out += fmt.Sprintf("\t%-20s\t%s\n", k, v)
	}
	out += "\nSignature algorithms:\n"
	for k, v := range signatureAlgs {
		out += fmt.Sprintf("\t%s\t%s\n", k, v)
	}
	out += "\nContent encryption algorithms:\n"
	for k, v := range contentEncryptionAlgs {
		out += fmt.Sprintf("\t%-10s\t%s\n", k, v)
	}
	return out
}

func getHeader(token *jwt.JSONWebToken, key string) (value string, err error) {
	for _, header := range token.Headers {
		if len(header.ExtraHeaders) > 0 {
			for k, v := range header.ExtraHeaders {
				if string(k) == key {
					value = fmt.Sprintf("%v", v)
					return
				}
			}
		}
	}
	err = errors.New("Key not found in headers")
	return
}

func loadPrivateKeyFile(path string) (interface{}, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadPrivateKey(data)
}

func loadPrivateKey(data []byte) (interface{}, error) {
	input := data

	block, _ := pem.Decode(data)
	if block != nil {
		input = block.Bytes
	}

	var priv interface{}
	priv, err0 := x509.ParsePKCS1PrivateKey(input)
	if err0 == nil {
		return priv, nil
	}

	priv, err1 := x509.ParsePKCS8PrivateKey(input)
	if err1 == nil {
		return priv, nil
	}

	priv, err2 := x509.ParseECPrivateKey(input)
	if err2 == nil {
		return priv, nil
	}

	return nil, fmt.Errorf("Unable to load private key")
}

func encode(obj interface{}) (outStr string, err error) {
	jsonEncoded, err := json.Marshal(obj)
	if err != nil {
		return
	}
	outStr = base64.RawURLEncoding.EncodeToString(jsonEncoded)
	return
}

func doJWS(reader *bufio.Reader, header string, signAlg string, signSecret []byte, signKey string, encAlg string, encSecret []byte, encKey []byte, keyAlg string, recreate bool) (outBuf []byte, err error) {
	var tokenSignature []byte
	var encodedPayload, encodedHeader string

	token := orderedmap.New()
	token.SetEscapeHTML(false)
	tokenHeader := orderedmap.New()
	tokenHeader.SetEscapeHTML(false)
	tokenPayload := orderedmap.New()
	tokenPayload.SetEscapeHTML(false)

	// If header is not set, we expect a full token object with header, payload and signature
	if header == "" {
		if err = json.NewDecoder(reader).Decode(&token); err != nil {
			return
		}
		if _, ok := token.Get("header"); !ok {
			// Create a default header
			if signAlg == "" {
				return nil, fmt.Errorf("Missing sign-alg")
			}
			tokenHeader.Set("alg", signAlg)
			tokenHeader.Set("typ", "JWT")
		} else {
			// Take the header from the given token
			curHeader, ok := token.Get("header")
			if !ok {
				return nil, fmt.Errorf("Could not get header")
			}
			bla := curHeader.(orderedmap.OrderedMap)
			tokenHeader = &bla
			if _, ok := tokenHeader.Get("alg"); ok {
				// Only overwrite if signAlg was not provided
				if signAlg == "" {
					signAlgRaw, ok := tokenHeader.Get("alg")
					if !ok {
						return nil, fmt.Errorf("Could not get alg from token header")
					}
					signAlg = signAlgRaw.(string)
				}
			}
		}
		payload, ok := token.Get("payload")
		if !ok {
			return nil, fmt.Errorf("Could not get payload")
		}
		more := payload.(orderedmap.OrderedMap)
		tokenPayload = &more

		// Take the signature from the given token
		if _, ok := token.Get("signature"); ok {
			asd, ok := token.Get("signature")
			if !ok {
				return nil, fmt.Errorf("Could not get signature")
			}
			xxx := asd.(string)
			tokenSignature = []byte(xxx)
		}

	} else {
		if err = json.NewDecoder(reader).Decode(&tokenPayload); err != nil {
			return
		}
		if err = json.Unmarshal([]byte(header), &tokenHeader); err != nil {
			return
		}
		if val, ok := tokenHeader.Get("alg"); ok {
			signAlg = fmt.Sprintf("%v", val)
		}
	}

	if len(tokenHeader.Keys()) < 1 || len(tokenPayload.Keys()) < 1 {
		err = fmt.Errorf("Token header or payload not set")
		return
	}

	if recreate {
		// Just create a new token based on the given JSON
		encodedHeader, err = encode(tokenHeader)
		if err != nil {
			return
		}
		encodedPayload, err = encode(tokenPayload)
		if err != nil {
			return
		}
		noneToken := fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, tokenSignature)
		outBuf = []byte(noneToken)
		return
	}

	if signAlg != "" && strings.ToLower(signAlg) == "none" {
		// Create a static token header with the given payload in the proper format
		tokenHeader.Set("alg", "none")
		encodedHeader, err = encode(tokenHeader)
		if err != nil {
			return
		}
		encodedPayload, err = encode(tokenPayload)
		if err != nil {
			return
		}
		noneToken := fmt.Sprintf("%s.%s.", encodedHeader, encodedPayload)
		outBuf = []byte(noneToken)
		return
	}

	var signer jose.Signer
	var encrypter jose.Encrypter

	if len(signSecret) > 0 || len(signKey) > 0 {
		// We do want to sign
		var privateKey interface{}
		privateKey, err = loadPrivateKeyFile(signKey)
		if err != nil {
			return
		}
		key := jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(signAlg),
			Key:       privateKey,
		}
		var opts jose.SignerOptions
		opts = jose.SignerOptions{EmbedJWK: false}
		opts.ExtraHeaders = make(map[jose.HeaderKey]interface{})
		// Copy attributes from given header into target header
		for _, k := range tokenHeader.Keys() {
			v, _ := tokenHeader.Get(k)
			opts.WithHeader(jose.HeaderKey(k), v)
		}
		var sig jose.Signer
		sig, err = jose.NewSigner(key, (&opts).WithType("JWT").WithContentType("JWT"))
		if err != nil {
			return
		}
		signer = sig
	}

	if encAlg != "" {
		var enc jose.Encrypter
		enc, err = jose.NewEncrypter(jose.ContentEncryption(encAlg), jose.Recipient{
			Algorithm: jose.KeyAlgorithm(keyAlg),
			Key:       encKey,
		}, &jose.EncrypterOptions{
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				jose.HeaderContentType: jose.ContentType("JWT"),
			},
		})
		if err != nil {
			return
		}
		encrypter = enc
	}

	var processedToken string
	if signer != nil && encrypter == nil {
		processedToken, err = jwt.Signed(signer).Claims(tokenPayload).Serialize()
		if err != nil {
			return
		}
	}
	if signer != nil && encrypter != nil {
		processedToken, err = jwt.SignedAndEncrypted(signer, encrypter).Claims(tokenPayload).Serialize()
		if err != nil {
			return
		}
	}
	if signer == nil && encrypter != nil {
		processedToken, err = jwt.Encrypted(encrypter).Claims(tokenPayload).Serialize()
		if err != nil {
			return
		}
	}
	outBuf = []byte(processedToken)
	return
}

func undoJWS(reader io.Reader, verify bool, secret []byte) (header, payload *orderedmap.OrderedMap, signature string, err error) {
	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(reader)

	parts := strings.Split(inBuf.String(), ".")
	if len(parts) < 1 || len(parts) > 3 {
		err = fmt.Errorf("Tokens must have at least one part and at max three parts")
		return
	}
	tokenHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return
	}
	err = json.Unmarshal(tokenHeader, &header)
	if err != nil {
		return
	}

	// Keep encoded signature
	if len(parts) > 2 {
		signature = strings.TrimSpace(parts[2])
	}

	token, err := jwt.ParseSigned(inBuf.String(), allowedSignatureAlgorithms())
	if err != nil {
		return
	}

	tokenCty, err := getHeader(token, "cty")
	if err == nil {
		// cty header found
		if tokenCty == "JWT" {
			// Nested token: https://tools.ietf.org/html/draft-yusef-oauth-nested-jwt-03
			err = errors.New("Nested tokens are currently not supported")
			//return
		}
	}

	if verify {
		// if err = token.Claims(secret, &payload); err != nil {
		// 	return
		// }
		//payloadTemp, err := token.Verify(secret)
	} else {
		if err = token.UnsafeClaimsWithoutVerification(&payload); err != nil {
			return
		}
	}
	return
}

func undoSignedJWE(reader io.Reader, verify bool, secret []byte) (header, payload *orderedmap.OrderedMap, signature string, err error) {
	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(reader)
	//var token *jwt.JSONWebToken
	//token, err = jwt.ParseSignedAndEncrypted(inBuf.String())
	//if err != nil {
	//return
	//}
	return
}

// NewPluginJwt creates a new JWT (RFC 7519) plugin.
func NewPluginJwt() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "jwt"
	p.Aliases = []string{".jwt"}
	p.Category = "formatters"
	p.Description = "Encode and decode JSON Web Tokens (JWT) (RFC 7519)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		// Encoding flags.
		flags.Bool("list", false, "list supported algorithms")
		flags.String("header", "", "token header")
		flags.String("sign-alg", "", "signature algorithm")
		flags.String("sign-secret", "", "signature secret")
		flags.String("sign-keyfile", "", "signature key file")
		flags.String("enc-alg", "", "encryption algorithm")
		flags.String("enc-secret", "", "encryption secret")
		flags.String("enc-keyfile", "", "encryption key file")
		flags.String("key-alg", "", "key management algorithm")
		flags.Bool("r", false, "recreate the token, keep the given signature")
		// Decoding flags.
		flags.Bool("verify", false, "verify signature")
		flags.String("secret", "", "secret key")
		flags.String("key", "", "key file")
		flags.Bool("decrypt", false, "decrypt JWE token")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		if helpers.IsBoolFlag(flags, "list") {
			_, err := io.WriteString(w, listAlgs())
			return err
		}
		out, err := doJWS(
			bufio.NewReader(r),
			helpers.StringFlag(flags, "header"),
			helpers.StringFlag(flags, "sign-alg"),
			[]byte(helpers.StringFlag(flags, "sign-secret")),
			helpers.StringFlag(flags, "sign-keyfile"),
			helpers.StringFlag(flags, "enc-alg"),
			[]byte(helpers.StringFlag(flags, "enc-secret")),
			[]byte(helpers.StringFlag(flags, "enc-keyfile")),
			helpers.StringFlag(flags, "key-alg"),
			helpers.IsBoolFlag(flags, "r"),
		)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		verify := helpers.IsBoolFlag(flags, "verify")
		secret := []byte(helpers.StringFlag(flags, "secret"))

		var (
			header, payload *orderedmap.OrderedMap
			signature       string
			err             error
		)
		if helpers.IsBoolFlag(flags, "decrypt") {
			header, payload, signature, err = undoSignedJWE(r, verify, secret)
		} else {
			header, payload, signature, err = undoJWS(r, verify, secret)
		}
		if err != nil {
			return err
		}
		if header == nil || payload == nil || len(header.Keys()) == 0 || len(payload.Keys()) == 0 {
			return nil
		}

		outObj := orderedmap.New()
		outObj.SetEscapeHTML(false)
		outObj.Set("header", header)
		outObj.Set("payload", payload)
		outObj.Set("signature", signature)

		out, err := prettyEncodeJSON(outObj)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	return p
}

// Helper functions

func prettyEncodeJSON(data interface{}) (outBuf []byte, err error) {
	var outBufWriter bytes.Buffer
	encoder := json.NewEncoder(&outBufWriter)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(data)
	if err != nil {
		return
	}
	outBuf = outBufWriter.Bytes()
	outBuf = bytes.TrimSuffix(outBuf, []byte("\n"))
	return
}
