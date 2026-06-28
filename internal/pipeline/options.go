package pipeline

import (
	"flag"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/internal/plugins"
)

// Option describes a single configurable plugin flag for UI rendering.
type Option struct {
	Name        string
	Label       string
	Usage       string
	Description string
	Default     string
	IsBool      bool
	Kind        string
	Choices     []string
	Secret      bool
	Multiline   bool
	HelpLabel   string
	HelpURL     string
}

// PluginOptions returns the configurable options (flags) of a plugin, or nil if
// it has none. Bool flags are reported with IsBool so the UI can render a
// checkbox instead of a text entry.
func PluginOptions(name string) []Option {
	p, _, ok := plugins.Resolve(name)
	if !ok || p.RegisterFlags == nil {
		return nil
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	p.RegisterFlags(fs)

	var opts []Option
	fs.VisitAll(func(f *flag.Flag) {
		_, isBool := f.Value.(interface{ IsBoolFlag() bool })
		kind := optionKind(p.Name, f.Name, f.DefValue, isBool)
		opts = append(opts, Option{
			Name:        f.Name,
			Label:       optionLabel(p.Name, f.Name),
			Usage:       f.Usage,
			Description: optionDescription(p.Name, f.Name, f.Usage),
			Default:     f.DefValue,
			IsBool:      isBool,
			Kind:        kind,
			Choices:     optionChoices(p.Name, f.Name),
			Secret:      isSecretOption(f.Name),
			Multiline:   optionMultiline(p.Name, f.Name),
			HelpLabel:   optionHelpLabel(p.Name, f.Name),
			HelpURL:     optionHelpURL(p.Name, f.Name),
		})
	})
	return opts
}

func optionLabel(plugin, name string) string {
	switch plugin + ":" + name {
	case "base32:no-pad":
		return "No padding"
	case "base64:url":
		return "URL-safe alphabet"
	case "base64:raw":
		return "Raw output"
	case "base64:strict":
		return "Strict decoding"
	case "base32:hex":
		return "Extended hex alphabet"
	case "blake3:derive-key":
		return "Derive key"
	case "brotli:lgwin":
		return "Window size"
	case "certCloner:ca-cert":
		return "CA certificate"
	case "certCloner:ca-key":
		return "CA private key"
	case "certCloner:o":
		return "Output prefix"
	case "certCloner:sig-alg":
		return "Signature algorithm"
	case "csv:in":
		return "Input format"
	case "csv:out":
		return "Output format"
	case "jq:no-color", "json:no-color":
		return "No color"
	case "jq:q":
		return "Query"
	case "jwt:decrypt":
		return "Decrypt JWE"
	case "jwt:enc-alg":
		return "Encryption algorithm"
	case "jwt:enc-keyfile":
		return "Encryption key file"
	case "jwt:enc-secret":
		return "Encryption secret"
	case "jwt:header":
		return "Token header"
	case "jwt:key":
		return "Verification key file"
	case "jwt:key-alg":
		return "Key management algorithm"
	case "jwt:list":
		return "List algorithms"
	case "jwt:r":
		return "Recreate token, keep signature"
	case "jwt:secret":
		return "Verification secret"
	case "jwt:sign-alg":
		return "Signing algorithm"
	case "jwt:sign-keyfile":
		return "Signing key file"
	case "jwt:sign-secret":
		return "Signing secret"
	case "jwt:verify":
		return "Verify signature"
	case "lzw:lit-width":
		return "Literal width"
	case "pem:cert":
		return "Certificate"
	case "pem:type":
		return "PEM type"
	case "qr:size":
		return "Image size"
	case "regex:re":
		return "Regular expression"
	case "saml:url":
		return "URL encoding"
	case "scrypt:len":
		return "Output length"
	case "scrypt:p":
		return "Parallelization"
	case "scrypt:r":
		return "Block size"
	case "sign:pub":
		return "Public key"
	case "sign:sig":
		return "Signature"
	case "strconv:ctrl":
		return "Control characters only"
	case "timestamp:utc":
		return "UTC output"
	case "uuid:gen":
		return "Generate UUID"
	case "unicode:big":
		return "Big endian"
	case "unicode:bom":
		return "BOM handling"
	case "unicode:encoding":
		return "Text encoding"
	case "unicode-normalize:form":
		return "Normalization form"
	case "ascii:mode":
		return "Non-ASCII handling"
	case "add:value":
		return "Add value"
	case "sub:value":
		return "Subtract value"
	case "xor:value":
		return "XOR value"
	default:
		return prettyOptionLabel(name)
	}
}

func optionDescription(plugin, name, usage string) string {
	switch plugin + ":" + name {
	case "add:value", "sub:value", "xor:value":
		return "Byte value as decimal, hex such as 0x2a, or a single character."
	case "aes:aad", "chacha20poly1305:aad":
		return "Additional authenticated data required for GCM or AEAD verification."
	case "aes:iv":
		return "Initialization vector for CBC or CTR mode. Accepts hex or Base64."
	case "aes:key":
		return "AES key as hex or Base64. Must be 16, 24, or 32 bytes."
	case "aes:mode":
		return "AES mode to use: GCM, CBC, or CTR."
	case "aes:nonce":
		return "GCM nonce as hex or Base64."
	case "ascii:mode":
		return "How to handle non-ASCII bytes."
	case "base32:hex":
		return "Use the RFC 4648 extended hex alphabet."
	case "base32:no-pad":
		return "Omit Base32 padding characters."
	case "base64:raw":
		return "Encode without Base64 padding."
	case "base64:strict":
		return "Decode only standard Base64."
	case "base64:url":
		return "Use the URL-safe Base64 alphabet."
	case "bcrypt:cost":
		return "bcrypt work factor."
	case "blake2b:key", "blake2s:key", "blake2x:key", "blake3:key", "hmac:key":
		return "Key for MAC or keyed hashing."
	case "blake2b:len", "blake2s:len", "blake2x:len", "blake3:length":
		return "Number of digest bytes to output."
	case "blake3:context":
		return "Context string for BLAKE3 key derivation."
	case "blake3:derive-key":
		return "Key material for BLAKE3 key derivation."
	case "brotli:level", "bzip2:level", "flate:level", "gzip:level", "zlib:level":
		return "Compression level."
	case "brotli:lgwin":
		return "Brotli sliding window size."
	case "certCloner:ca-cert":
		return "CA certificate in PEM format. May include the CA key."
	case "certCloner:ca-key":
		return "CA private key in PEM format."
	case "certCloner:o":
		return "Write cloned certificate files using this path prefix."
	case "certCloner:sig-alg":
		return "Signature hash for the cloned certificate."
	case "chacha20poly1305:key":
		return "ChaCha20-Poly1305 key as hex or Base64. Must be 32 bytes."
	case "chacha20poly1305:nonce":
		return "ChaCha20-Poly1305 nonce as hex or Base64. Must be 12 bytes."
	case "csv:in":
		return "Input delimiter or format."
	case "csv:out":
		return "Output delimiter or format."
	case "hmac:alg":
		return "Hash algorithm for HMAC."
	case "jq:no-color", "json:no-color":
		return "Disable ANSI color in formatted output."
	case "jq:q":
		return "jq filter expression to run against the JSON input."
	case "jq:plain":
		return "Print compact JSON instead of formatted output."
	case "jwk:public":
		return "Output public JWK material when possible."
	case "jwk:thumbprint":
		return "Output RFC 7638 SHA-256 thumbprints."
	case "jwt:decrypt":
		return "Decrypt a JWE token."
	case "jwt:enc-alg":
		return "Content encryption algorithm for creating JWE tokens."
	case "jwt:enc-keyfile":
		return "Encryption key file for creating JWE tokens."
	case "jwt:enc-secret":
		return "Encryption secret for creating JWE tokens."
	case "jwt:header":
		return "JSON header to use when creating a token."
	case "jwt:key":
		return "Key file used when verifying a token."
	case "jwt:key-alg":
		return "Key management algorithm for creating JWE tokens."
	case "jwt:list":
		return "Show supported JWT algorithms."
	case "jwt:r":
		return "Recreate the token with modified JSON while keeping the original signature."
	case "jwt:secret":
		return "Secret used when verifying a token."
	case "jwt:sign-alg":
		return "Signature algorithm for creating JWS tokens."
	case "jwt:sign-keyfile":
		return "Private key file used to sign a token."
	case "jwt:sign-secret":
		return "Shared secret used to sign a token."
	case "jwt:verify":
		return "Verify the token signature while decoding."
	case "lzw:lit-width":
		return "Number of bits used for literal codes."
	case "lzw:order":
		return "Bit order for LZW data."
	case "pem:cert":
		return "Create PEM output from certificate bytes."
	case "pem:headers":
		return "PEM headers as a JSON object."
	case "pem:type":
		return "PEM block type."
	case "qr:size":
		return "Generated QR image size in pixels."
	case "regex:all":
		return "Return all matches instead of the first match."
	case "regex:group":
		return "Capture group to extract."
	case "regex:re":
		return "Regular expression to match."
	case "regex:replace":
		return "Replacement text. When set, matching text is replaced."
	case "saml:deflate":
		return "Use raw DEFLATE compression for SAML redirect payloads."
	case "saml:plain":
		return "Skip DEFLATE detection when decoding."
	case "saml:url":
		return "URL-escape encoded output or URL-unescape input."
	case "scrypt:cost":
		return "CPU and memory cost parameter."
	case "scrypt:len":
		return "Number of key bytes to output."
	case "scrypt:p":
		return "scrypt parallelization parameter."
	case "scrypt:r":
		return "scrypt block size parameter."
	case "scrypt:salt":
		return "Salt as a hex string."
	case "sign:alg":
		return "Signature algorithm."
	case "sign:key":
		return "Private key for signing. Accepts PEM path or hex/Base64 Ed25519 key material."
	case "sign:pub":
		return "Public key for verification. Accepts PEM path or hex/Base64 Ed25519 key material."
	case "sign:sig":
		return "Signature to verify, as hex, Base64, or a file path."
	case "strconv:ctrl":
		return "Escape only control characters."
	case "strconv:graph":
		return "Escape printable characters using Go graph escapes."
	case "timestamp:layout":
		return "Go time layout for formatting or parsing time strings."
	case "timestamp:unit":
		return "Timestamp unit."
	case "timestamp:utc":
		return "Format parsed times in UTC."
	case "unicode-normalize:form":
		return "Unicode normalization form."
	case "unicode:big":
		return "Use big-endian byte order."
	case "unicode:bom":
		return "How to handle byte order marks."
	case "unicode:encoding":
		return "Text encoding for conversion."
	case "uuid:gen":
		return "Generate a random UUID v4."
	case "uuid:info":
		return "Show UUID version, variant, and bytes."
	default:
		return usage
	}
}

func prettyOptionLabel(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		parts[i] = prettyOptionPart(part)
	}
	return strings.Join(parts, " ")
}

func prettyOptionPart(part string) string {
	switch strings.ToLower(part) {
	case "aad":
		return "AAD"
	case "alg":
		return "Algorithm"
	case "bom":
		return "BOM"
	case "ca":
		return "CA"
	case "gcm":
		return "GCM"
	case "iv":
		return "IV"
	case "jwe":
		return "JWE"
	case "jwk":
		return "JWK"
	case "jwt":
		return "JWT"
	case "len":
		return "Length"
	case "mac":
		return "MAC"
	case "pem":
		return "PEM"
	case "pub":
		return "Public"
	case "qr":
		return "QR"
	case "re":
		return "Regex"
	case "sig":
		return "Signature"
	case "url":
		return "URL"
	case "utc":
		return "UTC"
	default:
		if part == "" {
			return ""
		}
		return strings.ToUpper(part[:1]) + part[1:]
	}
}

func optionMultiline(plugin, name string) bool {
	return plugin == "jq" && name == "q"
}

func optionHelpLabel(plugin, name string) string {
	switch plugin + ":" + name {
	case "jq:q":
		return "jq syntax"
	default:
		return ""
	}
}

func optionHelpURL(plugin, name string) string {
	switch plugin + ":" + name {
	case "jq:q":
		return "https://jqlang.github.io/jq/manual/"
	default:
		return ""
	}
}

func optionKind(plugin, name, def string, isBool bool) string {
	switch {
	case isBool:
		return "bool"
	case len(optionChoices(plugin, name)) > 0:
		return "select"
	case isSecretOption(name):
		return "secret"
	case isNumberDefault(def):
		return "number"
	default:
		return "text"
	}
}

func optionChoices(plugin, name string) []string {
	switch plugin + ":" + name {
	case "timestamp:unit":
		return []string{"auto", "s", "ms", "us", "ns"}
	case "unicode:bom":
		return []string{"ignore", "use", "expect"}
	case "unicode:encoding":
		return []string{
			"utf8",
			"utf16le",
			"utf16be",
			"utf32le",
			"utf32be",
			"latin1",
			"windows1252",
			"shiftjis",
			"eucjp",
			"gbk",
			"gb18030",
			"big5",
			"euckr",
			"koi8r",
		}
	case "unicode-normalize:form":
		return []string{"nfc", "nfd", "nfkc", "nfkd"}
	case "ascii:mode":
		return []string{"strict", "replace", "strip", "escape"}
	case "hmac:alg":
		return []string{"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3-256", "sha3-512"}
	case "lzw:order":
		return []string{"0", "1"}
	case "csv:in":
		return []string{"csv", "tsv", "semicolon"}
	case "csv:out":
		return []string{"table", "csv", "tsv", "semicolon"}
	case "aes:mode":
		return []string{"gcm", "cbc", "ctr"}
	case "sign:alg":
		return []string{"ed25519", "rsa-pss", "ecdsa"}
	default:
		return nil
	}
}

func isSecretOption(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "key") ||
		strings.Contains(name, "secret") ||
		strings.Contains(name, "pass") ||
		strings.Contains(name, "token")
}

func isNumberDefault(def string) bool {
	if def == "" {
		return false
	}
	_, err := strconv.Atoi(def)
	return err == nil
}
