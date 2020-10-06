package formatters

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/TylerBrock/colorjson"
	"github.com/takeshixx/deen/pkg/types"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
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

func encode(obj map[string]interface{}) (outStr string, err error) {
	jsonEncoded, err := json.Marshal(obj)
	if err != nil {
		return
	}
	outStr = base64.RawURLEncoding.EncodeToString(jsonEncoded)
	return
}

func doJWS(reader *bufio.Reader, header string, signAlg string, signSecret []byte, signKey []byte, encAlg string, encSecret []byte, encKey []byte, keyAlg string) (outBuf []byte, err error) {
	var token, tokenPayload, tokenHeader map[string]interface{}
	var encodedPayload, encodedHeader string

	// If header is not set, we expect a full token object with header, payload and signature
	if header == "" {
		if err = json.NewDecoder(reader).Decode(&token); err != nil {
			return
		}
		if _, ok := token["header"]; !ok {
			// Create a default headerg
			tokenHeader = make(map[string]interface{})
			tokenHeader["alg"] = signAlg
			tokenHeader["typ"] = "JWT"
		} else {
			header := token["header"]
			tokenHeader = header.(map[string]interface{})
			if _, ok := tokenHeader["alg"]; ok {
				signAlg = tokenHeader["alg"].(string)
			}
		}
		if _, ok := token["payload"]; !ok {
			return nil, fmt.Errorf("No payload found in input")
		}
		payload := token["payload"]
		tokenPayload = payload.(map[string]interface{})
	} else {
		if err = json.NewDecoder(reader).Decode(&tokenPayload); err != nil {
			return
		}
		if err = json.Unmarshal([]byte(header), &tokenHeader); err != nil {
			return
		}
		if val, ok := tokenHeader["alg"]; ok {
			signAlg = fmt.Sprintf("%v", val)
		}
	}

	if len(tokenHeader) < 1 || len(tokenPayload) < 1 {
		err = fmt.Errorf("Token header or payload not set")
		return
	}

	if signAlg != "" && strings.ToLower(signAlg) == "none" {
		// Create a static token header with the given payload in the proper format
		tokenHeader["alg"] = "none"
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
		key := jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(signAlg),
			Key:       signSecret,
		}
		var sig jose.Signer
		sig, err = jose.NewSigner(key, (&jose.SignerOptions{EmbedJWK: false}).WithType("JWT").WithContentType("JWT"))
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
		processedToken, err = jwt.Signed(signer).Claims(tokenPayload).CompactSerialize()
		if err != nil {
			return
		}
	}
	if signer != nil && encrypter != nil {
		processedToken, err = jwt.SignedAndEncrypted(signer, encrypter).Claims(tokenPayload).CompactSerialize()
		if err != nil {
			return
		}
	}
	if signer == nil && encrypter != nil {
		processedToken, err = jwt.Encrypted(encrypter).Claims(tokenPayload).CompactSerialize()
		if err != nil {
			return
		}
	}
	outBuf = []byte(processedToken)
	return
}

func printSerializedToken(token string) (outBuf []byte, err error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(token), &raw); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", raw)
	return
}

func undoJWS(reader io.Reader, verify bool, secret []byte) (header, payload map[string]interface{}, signature string, err error) {
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

	token, err := jwt.ParseSigned(inBuf.String())
	if err != nil {
		return
	}

	tokenCty, err := getHeader(token, "cty")
	if err == nil {
		// cty header found
		if tokenCty == "JWT" {
			// Nested token: https://tools.ietf.org/html/draft-yusef-oauth-nested-jwt-03
			err = errors.New("Nested tokens are currently not supported")
			return
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

func undoJWE(reader io.Reader, secret []byte) (header, payload map[string]interface{}, signature string, err error) {
	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(reader)
	//var token *jwt.JSONWebToken
	//token, err = jwt.ParseEncrypted(inBuf.String())
	//if err != nil {
	//return
	//}
	return
}

func undoSignedJWE(reader io.Reader, verify bool, secret []byte) (header, payload map[string]interface{}, signature string, err error) {
	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(reader)
	//var token *jwt.JSONWebToken
	//token, err = jwt.ParseSignedAndEncrypted(inBuf.String())
	//if err != nil {
	//return
	//}
	return
}

// NewPluginJwt creates a new PluginJwt object
func NewPluginJwt() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "jwt"
	p.Aliases = []string{".jwt"}
	p.Category = "formatters"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var outBuf bytes.Buffer
		var err error

		return outBuf.Bytes(), err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) (outBuf []byte, err error) {
		listFlag := flags.Lookup("list")
		if listCmd, err := strconv.ParseBool(listFlag.Value.String()); listCmd && err == nil {
			// List supported algoritms
			out := []byte(listAlgs())
			return out, err
		}

		var signAlg, signSecret, signKey, encAlg, encSecret, encKey, keyAlg, header string

		if signAlgFlag := flags.Lookup("sign-alg"); signAlgFlag != nil {
			signAlg = signAlgFlag.Value.String()
		}
		if signSecretFlag := flags.Lookup("sign-secret"); signSecretFlag != nil {
			signSecret = signSecretFlag.Value.String()
		}
		if signKeyFlag := flags.Lookup("sign-keyfile"); signKeyFlag != nil {
			signKey = signKeyFlag.Value.String()
		}
		if encAlgFlag := flags.Lookup("enc-alg"); encAlgFlag != nil {
			encAlg = encAlgFlag.Value.String()
		}
		if encSecretFlag := flags.Lookup("enc-secret"); encSecretFlag != nil {
			encSecret = encSecretFlag.Value.String()
		}
		if encKeyFlag := flags.Lookup("enc-keyfile"); encKeyFlag != nil {
			encKey = encKeyFlag.Value.String()
		}
		if keyAlgFlag := flags.Lookup("key-alg"); keyAlgFlag != nil {
			keyAlg = keyAlgFlag.Value.String()
		}
		if headerFlag := flags.Lookup("header"); headerFlag != nil {
			header = headerFlag.Value.String()
		}

		inBuf := bufio.NewReader(reader)

		// In case there is no input, print the help page
		if inBuf.Size() < 1 {
			flags.Usage()
			os.Exit(1)
		}

		return doJWS(inBuf, header, signAlg, []byte(signSecret), []byte(signKey), encAlg, []byte(encSecret), []byte(encKey), keyAlg)
	}
	p.UnprocessStreamFunc = func(reader io.Reader) ([]byte, error) {
		var secret, headerBuf, payloadBuf []byte
		var err error
		header, payload, signature, err := undoJWS(reader, false, secret)
		if err != nil {
			return nil, err
		}
		headerBuf, err = json.Marshal(header)
		if err != nil {
			return nil, err
		}
		payloadBuf, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		outStr := fmt.Sprintf("%s %s %s", string(headerBuf), string(payloadBuf), signature)
		outBuf := []byte(outStr)
		return outBuf, nil

		//return undoJWS(reader, false, secret)
	}
	p.UnprocessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) (outBuf []byte, err error) {
		verifyFlag := flags.Lookup("verify")
		verify, err := strconv.ParseBool(verifyFlag.Value.String())
		if err != nil {
			return
		}

		jweFlag := flags.Lookup("decrypt")
		isJWE, err := strconv.ParseBool(jweFlag.Value.String())
		if err != nil {
			return
		}

		secretFlag := flags.Lookup("secret")
		secret := []byte(secretFlag.Value.String())

		var header, payload map[string]interface{}
		var signature string

		if isJWE {
			header, payload, signature, err = undoSignedJWE(reader, verify, secret)
		} else if isJWE {
			header, payload, signature, err = undoJWE(reader, secret)
		} else {
			header, payload, signature, err = undoJWS(reader, verify, secret)
		}

		if err != nil {
			return
		}

		if len(header) == 0 || len(payload) == 0 {
			return
		}

		plainFlag := flags.Lookup("plain")
		plain, err := strconv.ParseBool(plainFlag.Value.String())

		noColorFlag := flags.Lookup("no-color")
		noColor, err := strconv.ParseBool(noColorFlag.Value.String())

		outObj := make(map[string]interface{})
		outObj["header"] = header
		outObj["payload"] = payload
		outObj["signature"] = signature

		if !plain {
			if noColor {
				outBuf, err = prettyEncodeJSON(outObj)
			} else {
				outBuf, err = prettyEncodeJSONColored(outObj)
			}
		} else {
			outBuf, err = json.Marshal(outObj)
		}
		return
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		if self.Unprocess {
			// Decoding
			flags.Usage = func() {
				fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
				fmt.Fprintf(os.Stderr, "Decode JSON Web Tokens (JWT) (RFC7519).\n\n")
				flags.PrintDefaults()
			}
			flags.Bool("verify", false, "verify signature")
			flags.String("secret", "", "secret key")
			flags.String("key", "", "key file")
			flags.Bool("decrypt", false, "decrypt JWE token")
			flags.Bool("plain", false, "print unformatted token")
			flags.Bool("no-color", false, "omit colors in formatted output")
			flags.Parse(args)
			return flags
		}
		// Encoding
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Encode JSON Web Tokens (JWT) (RFC7519).\n\n")
			flags.PrintDefaults()
		}
		flags.Bool("list", false, "list supported algorithms")
		flags.String("header", "", "token header")

		flags.String("sign-alg", "HS256", "signature algorithm")
		flags.String("sign-secret", "", "signature secret")
		flags.String("sign-keyfile", "", "signature key file")
		flags.String("enc-alg", "", "encryption algorithm")
		flags.String("enc-secret", "", "encryption secret")
		flags.String("enc-keyfile", "", "encryption key file")
		flags.String("key-alg", "", "key management algorithm")
		flags.Parse(args)
		return flags
	}
	return
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
	return
}

func prettyEncodeJSONColored(data interface{}) (outBuf []byte, err error) {
	f := colorjson.NewFormatter()
	f.Indent = 4
	return f.Marshal(data)
}
