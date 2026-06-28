package plugins

import "strings"

// Reference points users to background material for a plugin or family.
type Reference struct {
	Label string
	URL   string
}

// Example gives a small input/output pair for a catalog entry.
type Example struct {
	Label  string
	Input  string
	Output string
}

// UIPluginInfo is catalog metadata shaped for human-facing interfaces.
type UIPluginInfo struct {
	Name        string
	Label       string
	Category    string
	Aliases     []string
	Description string
	UseFor      string
	CanDecode   bool
	References  []Reference
	Examples    []Example
}

type catalogCopy struct {
	Description string
	UseFor      string
	References  []Reference
	Examples    []Example
}

// CategoryLabel returns the display label used by graphical catalog views.
func CategoryLabel(category string) string {
	switch category {
	case "codecs":
		return "Codecs"
	case "compressions":
		return "Compressions"
	case "hashs":
		return "Hashes"
	case "formatters":
		return "Formatters"
	case "misc":
		return "Misc"
	case "arithmetic":
		return "Arithmetic"
	default:
		return category
	}
}

// CategorySelectLabel returns placeholder text for category-specific transform
// pickers.
func CategorySelectLabel(category string) string {
	switch category {
	case "codecs":
		return "Select codec"
	case "compressions":
		return "Select compression"
	case "hashs":
		return "Select hash"
	case "formatters":
		return "Select formatter"
	case "misc":
		return "Select misc tool"
	case "arithmetic":
		return "Select arithmetic"
	default:
		return "Select transformer"
	}
}

// PluginLabel returns the user-facing display name for a plugin while keeping
// command IDs stable elsewhere.
func PluginLabel(name string) string {
	if label, ok := pluginLabels[name]; ok {
		return label
	}
	return name
}

var pluginLabels = map[string]string{
	"html":              "HTML",
	"url":               "URL",
	"unicode":           "Unicode",
	"unicode-inspect":   "Unicode Inspect",
	"unicode-normalize": "Unicode Normalize",
	"ascii":             "ASCII",
	"pem":               "PEM",
	"quoted-printable":  "Quoted-Printable",
	"rot13":             "ROT13",
	"hmac":              "HMAC",
	"json":              "JSON",
	"xml":               "XML",
	"json2xml":          "JSON to XML",
	"toml":              "TOML",
	"jwt":               "JWT",
	"jwk":               "JWK",
	"jq":                "jq",
	"protobuf":          "Protocol Buffers",
	"msgpack":           "MessagePack",
	"cbor":              "CBOR",
	"yaml":              "YAML",
	"csv":               "CSV",
	"qr":                "QR Code",
	"saml":              "SAML",
	"asn1":              "ASN.1",
	"dns":               "DNS",
	"uuid":              "UUID",
	"aes":               "AES",
	"chacha20poly1305":  "ChaCha20-Poly1305",
	"certCloner":        "Certificate Cloner",
	"certPrinter":       "Certificate Printer",
	"crc32":             "CRC-32",
	"crc32c":            "CRC-32C",
	"crc32k":            "CRC-32 Koopman",
	"crc64":             "CRC-64 ISO",
	"crc64-ecma":        "CRC-64 ECMA",
	"fnv32":             "FNV-32",
	"fnv32a":            "FNV-32a",
	"fnv64":             "FNV-64",
	"fnv64a":            "FNV-64a",
	"fnv128":            "FNV-128",
	"fnv128a":           "FNV-128a",
	"ripemd160":         "RIPEMD-160",
	"blake2s":           "BLAKE2s",
	"blake2b":           "BLAKE2b",
	"blake2x":           "BLAKE2x",
	"blake3":            "BLAKE3",
	"sha1":              "SHA-1",
	"sha224":            "SHA-224",
	"sha256":            "SHA-256",
	"sha384":            "SHA-384",
	"sha512":            "SHA-512",
	"sha512-224":        "SHA-512/224",
	"sha512-256":        "SHA-512/256",
	"sha3-224":          "SHA3-224",
	"sha3-256":          "SHA3-256",
	"sha3-384":          "SHA3-384",
	"sha3-512":          "SHA3-512",
	"md4":               "MD4",
	"md5":               "MD5",
	"lzma":              "LZMA",
	"lzma2":             "LZMA2",
	"xor":               "XOR",
	"not":               "NOT",
}

var referenceSets = map[string][]Reference{
	"rfc4648": {
		{"RFC 4648", "https://www.rfc-editor.org/rfc/rfc4648"},
	},
	"html": {
		{"HTML Standard: character references", "https://html.spec.whatwg.org/multipage/named-characters.html"},
	},
	"url": {
		{"RFC 3986", "https://www.rfc-editor.org/rfc/rfc3986"},
	},
	"unicode": {
		{"Unicode Standard", "https://www.unicode.org/standard/standard.html"},
	},
	"pem": {
		{"RFC 7468", "https://www.rfc-editor.org/rfc/rfc7468"},
	},
	"quoted-printable": {
		{"RFC 2045", "https://www.rfc-editor.org/rfc/rfc2045"},
	},
	"gzip": {
		{"RFC 1952", "https://www.rfc-editor.org/rfc/rfc1952"},
	},
	"zlib": {
		{"RFC 1950", "https://www.rfc-editor.org/rfc/rfc1950"},
	},
	"deflate": {
		{"RFC 1951", "https://www.rfc-editor.org/rfc/rfc1951"},
	},
	"bzip2": {
		{"bzip2 format", "https://sourceware.org/bzip2/"},
	},
	"lzma": {
		{"LZMA SDK", "https://www.7-zip.org/sdk.html"},
	},
	"lzw": {
		{"Go compress/lzw", "https://pkg.go.dev/compress/lzw"},
	},
	"brotli": {
		{"RFC 7932", "https://www.rfc-editor.org/rfc/rfc7932"},
	},
	"zstd": {
		{"RFC 8878", "https://www.rfc-editor.org/rfc/rfc8878"},
	},
	"sha1": {
		{"FIPS 180-4", "https://csrc.nist.gov/pubs/fips/180-4/upd1/final"},
	},
	"sha2": {
		{"FIPS 180-4", "https://csrc.nist.gov/pubs/fips/180-4/upd1/final"},
	},
	"sha3": {
		{"FIPS 202", "https://csrc.nist.gov/pubs/fips/202/final"},
	},
	"md": {
		{"RFC 1321", "https://www.rfc-editor.org/rfc/rfc1321"},
	},
	"ripemd": {
		{"RIPEMD-160", "https://homes.esat.kuleuven.be/~bosselae/ripemd160.html"},
	},
	"blake2": {
		{"RFC 7693", "https://www.rfc-editor.org/rfc/rfc7693"},
	},
	"blake3": {
		{"BLAKE3", "https://github.com/BLAKE3-team/BLAKE3-specs"},
	},
	"bcrypt": {
		{"bcrypt paper", "https://www.usenix.org/legacy/events/usenix99/provos/provos_html/"},
	},
	"scrypt": {
		{"RFC 7914", "https://www.rfc-editor.org/rfc/rfc7914"},
	},
	"hmac": {
		{"RFC 2104", "https://www.rfc-editor.org/rfc/rfc2104"},
	},
	"adler": {
		{"RFC 1950", "https://www.rfc-editor.org/rfc/rfc1950"},
	},
	"crc": {
		{"A painless guide to CRC error detection algorithms", "https://www.ross.net/crc/download/crc_v3.txt"},
	},
	"fnv": {
		{"FNV hash", "https://www.ietf.org/archive/id/draft-eastlake-fnv-25.html"},
	},
	"json": {
		{"RFC 8259", "https://www.rfc-editor.org/rfc/rfc8259"},
	},
	"xml": {
		{"XML 1.0", "https://www.w3.org/TR/xml/"},
	},
	"toml": {
		{"TOML", "https://toml.io/"},
	},
	"jwt": {
		{"RFC 7519", "https://www.rfc-editor.org/rfc/rfc7519"},
	},
	"jwk": {
		{"RFC 7517", "https://www.rfc-editor.org/rfc/rfc7517"},
		{"RFC 7638", "https://www.rfc-editor.org/rfc/rfc7638"},
	},
	"jq": {
		{"jq manual", "https://jqlang.github.io/jq/manual/"},
	},
	"protobuf": {
		{"Protocol Buffers encoding", "https://protobuf.dev/programming-guides/encoding/"},
	},
	"msgpack": {
		{"MessagePack", "https://msgpack.org/"},
	},
	"cbor": {
		{"RFC 8949", "https://www.rfc-editor.org/rfc/rfc8949"},
	},
	"yaml": {
		{"YAML 1.2.2", "https://yaml.org/spec/1.2.2/"},
	},
	"csv": {
		{"RFC 4180", "https://www.rfc-editor.org/rfc/rfc4180"},
	},
	"saml": {
		{"SAML Bindings", "https://docs.oasis-open.org/security/saml/v2.0/saml-bindings-2.0-os.pdf"},
		{"SAML Core", "https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf"},
	},
	"timestamp": {
		{"Unix time", "https://en.wikipedia.org/wiki/Unix_time"},
		{"Go time layouts", "https://pkg.go.dev/time#pkg-constants"},
	},
	"x509": {
		{"RFC 5280", "https://www.rfc-editor.org/rfc/rfc5280"},
	},
	"asn1": {
		{"X.690 ASN.1 encoding rules", "https://www.itu.int/rec/T-REC-X.690"},
	},
	"uuid": {
		{"RFC 9562", "https://www.rfc-editor.org/rfc/rfc9562"},
	},
	"dns": {
		{"RFC 1035", "https://www.rfc-editor.org/rfc/rfc1035"},
	},
	"magic": {
		{"WHATWG MIME sniffing", "https://mimesniff.spec.whatwg.org/"},
	},
	"aes": {
		{"NIST SP 800-38D", "https://csrc.nist.gov/pubs/sp/800/38/d/final"},
	},
	"chacha": {
		{"RFC 8439", "https://www.rfc-editor.org/rfc/rfc8439"},
	},
	"sign": {
		{"RFC 8032", "https://www.rfc-editor.org/rfc/rfc8032"},
	},
}

var catalogCopyByName = map[string]catalogCopy{
	"base32": {
		"Encodes binary data as Base32 text and decodes Base32 back to bytes.",
		"Use it when data must travel through uppercase, text-only channels such as tokens, provisioning secrets, or DNS-adjacent workflows.",
		referenceSets["rfc4648"],
		nil,
	},
	"base64": {
		"Encodes bytes as Base64 text and decodes standard, raw, URL-safe, and raw URL-safe Base64.",
		"Use it for common API payloads, certificates, JWT sections, cookies, and other compact binary-to-text data.",
		referenceSets["rfc4648"],
		[]Example{{"Encode text", "test", "dGVzdA=="}},
	},
	"base85": {
		"Encodes bytes using the denser Ascii85/Base85 representation and decodes it again.",
		"Use it when you see compact printable data from PostScript/PDF-style tools or need less overhead than Base64.",
		[]Reference{{"Ascii85 overview", "https://en.wikipedia.org/wiki/Ascii85"}},
		nil,
	},
	"ascii": {
		"Converts UTF-8 text to ASCII with explicit handling for non-ASCII characters.",
		"Use strict mode to validate ASCII-only text, or replace/strip/escape mode when you need an ASCII-safe representation.",
		referenceSets["unicode"],
		[]Example{{"Escape non-ASCII text", "Café 😀", "Caf\\u00E9 \\U0001F600"}},
	},
	"hex": {
		"Converts bytes to hexadecimal text and parses hexadecimal text back to bytes.",
		"Use it for hashes, binary protocol fragments, keys, signatures, and quick byte-level inspection.",
		nil,
		[]Example{{"Encode bytes", "test", "74657374"}},
	},
	"url": {
		"Escapes and unescapes URL query text using percent encoding.",
		"Use it to inspect query parameters, callback URLs, webhooks, and payloads copied from browser or proxy traffic.",
		referenceSets["url"],
		[]Example{{"Decode parameter text", "hello%20world", "hello world"}},
	},
	"html": {
		"Escapes and unescapes HTML character entities.",
		"Use it when web output contains entities like ampersand escapes or numeric character references.",
		referenceSets["html"],
		nil,
	},
	"unicode": {
		"Transcodes text between UTF-8, UTF-16, UTF-32, and common legacy character sets.",
		"Use it when readable text is stored with a different byte encoding, byte order, BOM convention, or charset such as Windows-1252 or Shift-JIS.",
		referenceSets["unicode"],
		[]Example{{"UTF-16LE round trip", "deen", "6400650065006e00"}},
	},
	"unicode-inspect": {
		"Reports Unicode and text-encoding clues without changing the input bytes.",
		"Use it to identify BOMs, UTF-8 validity, likely UTF-16/UTF-32 byte order, and suspicious control characters.",
		referenceSets["unicode"],
		nil,
	},
	"unicode-normalize": {
		"Normalizes UTF-8 text to NFC, NFD, NFKC, or NFKD.",
		"Use it when two strings look the same but compare differently, or when compatibility characters need a canonical form.",
		referenceSets["unicode"],
		[]Example{{"Compose accent marks", "Cafe\u0301", "Caf\u00e9"}},
	},
	"strconv": {
		"Quotes and unquotes strings using Go-style escape sequences.",
		"Use it for debugging string literals, control characters, and copied Go or JSON-like escaped text.",
		[]Reference{{"Go strconv package", "https://pkg.go.dev/strconv"}},
		nil,
	},
	"pem": {
		"Wraps DER bytes into PEM blocks and unwraps PEM text back to DER bytes.",
		"Use it for certificates, keys, CSRs, and other ASN.1 material exchanged as PEM text.",
		referenceSets["pem"],
		nil,
	},
	"quoted-printable": {
		"Encodes mostly-readable text as quoted-printable and decodes it again.",
		"Use it for email bodies, MIME parts, and payloads where non-ASCII bytes are represented as equals escapes.",
		referenceSets["quoted-printable"],
		nil,
	},
	"rot13": {
		"Applies the ROT13 substitution cipher. Running it twice restores the original text.",
		"Use it for simple obfuscation, puzzle text, or legacy examples. It is not encryption.",
		[]Reference{{"ROT13", "https://en.wikipedia.org/wiki/ROT13"}},
		nil,
	},
	"flate": {
		"Compresses and decompresses raw DEFLATE streams.",
		"Use it for low-level compressed data where there is no gzip or zlib wrapper.",
		referenceSets["deflate"],
		nil,
	},
	"gzip": {
		"Compresses and decompresses gzip streams.",
		"Use it for HTTP content, log archives, command-line gzip data, and portable compressed blobs.",
		referenceSets["gzip"],
		nil,
	},
	"zlib": {
		"Compresses and decompresses zlib-wrapped DEFLATE streams.",
		"Use it for PNG-adjacent data, application protocols, or formats that carry zlib streams.",
		referenceSets["zlib"],
		nil,
	},
	"bzip2": {
		"Compresses and decompresses bzip2 data.",
		"Use it for older archives or data where bzip2's high compression ratio matters more than speed.",
		referenceSets["bzip2"],
		nil,
	},
	"lzma": {
		"Compresses and decompresses LZMA data.",
		"Use it for 7-Zip/XZ-adjacent workflows and compact archival data.",
		referenceSets["lzma"],
		nil,
	},
	"lzma2": {
		"Compresses and decompresses LZMA2 data.",
		"Use it for XZ/7-Zip-style data that uses the newer LZMA2 stream format.",
		referenceSets["lzma"],
		nil,
	},
	"lzw": {
		"Compresses and decompresses LZW data.",
		"Use it for legacy formats and protocol samples that still rely on LZW coding.",
		referenceSets["lzw"],
		nil,
	},
	"brotli": {
		"Compresses and decompresses Brotli data.",
		"Use it for modern web compression, HTTP responses, and compact static asset payloads.",
		referenceSets["brotli"],
		nil,
	},
	"zstd": {
		"Compresses and decompresses Zstandard data.",
		"Use it for modern high-speed compression, logs, backups, and large payloads.",
		referenceSets["zstd"],
		nil,
	},
	"hmac": {
		"Computes keyed message authentication codes with selectable hash algorithms.",
		"Use it to verify webhook signatures, signed API requests, and integrity checks that require a shared secret.",
		referenceSets["hmac"],
		nil,
	},
	"bcrypt": {
		"Derives password hashes using bcrypt.",
		"Use it for password-hash experiments and verification fixtures. It is intentionally slow and one-way.",
		referenceSets["bcrypt"],
		nil,
	},
	"scrypt": {
		"Derives memory-hard hashes using scrypt.",
		"Use it for password-derived keys and password-hash testing where memory cost matters.",
		referenceSets["scrypt"],
		nil,
	},
	"json": {
		"Formats or minifies JSON.",
		"Use it to make API responses readable, normalize JSON before comparing it, or compact JSON for transport.",
		referenceSets["json"],
		[]Example{{"Format object", `{"ok":true}`, "{\n    \"ok\": true\n}"}},
	},
	"xml": {
		"Formats or minifies XML.",
		"Use it for SOAP, SAML, configuration files, and XML API responses.",
		referenceSets["xml"],
		nil,
	},
	"json2xml": {
		"Converts JSON to XML and XML back to JSON.",
		"Use it when moving data between services or tooling that expect different structured formats.",
		[]Reference{referenceSets["json"][0], referenceSets["xml"][0]},
		nil,
	},
	"toml": {
		"Normalizes TOML and converts TOML to formatted JSON in decode mode.",
		"Use it to inspect config files such as pyproject.toml, Cargo manifests, and application settings.",
		referenceSets["toml"],
		[]Example{{"Format config", "name = \"deen\"", "name = \"deen\""}},
	},
	"jwt": {
		"Decodes JWTs into readable header and claim data.",
		"Use it to inspect authentication tokens, claim sets, algorithms, and expiration fields.",
		referenceSets["jwt"],
		[]Example{{"Decode token", "eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjMifQ.", "{ header, payload, signature }"}},
	},
	"jwk": {
		"Formats, compacts, publicizes and thumbprints JSON Web Keys and JSON Web Key Sets.",
		"Use it to inspect OAuth/OIDC key sets, normalize JWK JSON, strip private key material where possible, and compute SHA-256 JWK thumbprints.",
		referenceSets["jwk"],
		[]Example{{"Pretty-print symmetric key", `{"kty":"oct","kid":"hmac","k":"AyM1SysPpbyDfgZld3umTQ"}`, "{\n  \"kty\": \"oct\",\n  ...\n}"}},
	},
	"jq": {
		"Runs jq-style queries and filters against JSON input.",
		"Use it to extract fields, reshape responses, and test JSON filters without leaving deen.",
		referenceSets["jq"],
		nil,
	},
	"protobuf": {
		"Decodes schema-less Protocol Buffers wire data into a readable field listing.",
		"Use it to inspect raw protobuf payloads from logs, captures, or binary API traffic when you do not have the .proto schema handy.",
		referenceSets["protobuf"],
		nil,
	},
	"msgpack": {
		"Encodes JSON to MessagePack and decodes MessagePack back to formatted JSON.",
		"Use it for compact binary API payloads, caches, fixtures, and protocol samples that use MessagePack.",
		referenceSets["msgpack"],
		[]Example{{"JSON to MessagePack", `{"ok":true}`, "81a26f6bc3"}},
	},
	"cbor": {
		"Encodes JSON to CBOR and decodes CBOR back to formatted JSON.",
		"Use it for COSE/WebAuthn-adjacent data, IoT payloads, and compact binary structured values.",
		referenceSets["cbor"],
		[]Example{{"JSON to CBOR", `{"ok":true}`, "a1626f6bf5"}},
	},
	"yaml": {
		"Normalizes YAML and converts YAML to compact JSON in decode mode.",
		"Use it to inspect configuration files, CI manifests, Kubernetes snippets, and YAML API payloads.",
		referenceSets["yaml"],
		[]Example{{"Format mapping", "name: deen\nok: true", "name: deen\nok: true\n"}},
	},
	"csv": {
		"Formats CSV/TSV as readable tables and converts between delimiters.",
		"Use it to inspect copied spreadsheet data, logs, exports, TSV payloads, and comma/semicolon-delimited records.",
		referenceSets["csv"],
		[]Example{{"CSV to TSV", "name,ok\ndeen,true", "name\tok\ndeen\ttrue"}},
	},
	"qr": {
		"Encodes text as a QR PNG and decodes QR images back to text.",
		"Use it for QR payload inspection, fixture generation, tokens, URLs, and mobile handoff workflows.",
		nil,
		[]Example{{"Encode text", "deen", "PNG image containing QR code"}},
	},
	"saml": {
		"Decodes SAMLRequest and SAMLResponse payloads into readable XML and encodes XML back to SAML payloads.",
		"Use it to inspect SAML HTTP-POST and HTTP-Redirect binding values copied from forms, URLs, browser tools, or proxy logs.",
		referenceSets["saml"],
		[]Example{{"Decode payload", "SAMLRequest=...", "<AuthnRequest ...>"}},
	},
	"timestamp": {
		"Converts Unix timestamps to formatted times and formatted times back to Unix timestamps.",
		"Use it for logs, API payloads, JWT claims, database rows, and any workflow that mixes epoch seconds, milliseconds, microseconds, nanoseconds, and RFC3339-style strings.",
		referenceSets["timestamp"],
		[]Example{{"Milliseconds to time", "1700000000123", "2023-11-14T22:13:20.123Z"}},
	},
	"asn1": {
		"Parses ASN.1 DER or PEM input into a readable tag/length/value tree.",
		"Use it to inspect certificates, keys, CSRs, signatures, and other DER-encoded structures when you need the raw TLV layout.",
		referenceSets["asn1"],
		[]Example{
			{"Inspect sequence", "300302012a", "SEQUENCE\n    INTEGER = 42"},
			{"Inspect PEM certificate", "-----BEGIN CERTIFICATE-----...", ".pem -> asn1 prints the DER structure"},
		},
	},
	"dns": {
		"Encodes and decodes DNS names in wire format.",
		"Use it to inspect DNS labels inside packets, logs, binary fixtures, and protocol examples.",
		referenceSets["dns"],
		[]Example{{"Encode name", "www.example.com.", "03777777076578616d706c6503636f6d00"}},
	},
	"uuid": {
		"Generates UUID v4 values, formats raw UUID bytes, and inspects UUID version and variant.",
		"Use it for API identifiers, binary UUID fields, logs, fixtures, and checking UUID shape.",
		referenceSets["uuid"],
		[]Example{{"Inspect UUID", "550e8400-e29b-41d4-a716-446655440000", "version: 4\nvariant: RFC 4122"}},
	},
	"entropy": {
		"Analyzes byte entropy, unique byte count, and the most common byte values.",
		"Use it to inspect randomness, encoded/compressed/encrypted-looking data, and binary payload distribution.",
		nil,
		[]Example{{"Repeated bytes", "aaaa", "entropy: 0.0000 bits/byte\n0x61: 4 100.00%"}},
	},
	"magic": {
		"Detects common file types from magic bytes and MIME content sniffing.",
		"Use it to quickly identify dropped files, binary blobs, compressed data, images, PDFs, and executables.",
		referenceSets["magic"],
		[]Example{{"Detect PNG", "89504e470d0a1a0a", "type: PNG image\nmime: image/png"}},
	},
	"regex": {
		"Extracts regular expression matches or replaces matched text.",
		"Use it to pull tokens, IDs, headers, URLs, and fields out of text before feeding another transform.",
		nil,
		[]Example{{"Extract IDs", "id=123 id=456", "123\n456"}},
	},
	"aes": {
		"Encrypts and decrypts with AES-GCM, AES-CBC, or AES-CTR.",
		"Use it for lab payloads, fixtures, and protocol samples where you have the key and nonce or IV. GCM supports configurable tag lengths; CBC supports PKCS#7 and unpadded block data.",
		referenceSets["aes"],
		[]Example{{"AES-GCM", "secret", "ciphertext plus authentication tag"}},
	},
	"chacha20poly1305": {
		"Encrypts and decrypts with ChaCha20-Poly1305.",
		"Use it for modern AEAD payloads where you have the 32-byte key and 12-byte nonce.",
		referenceSets["chacha"],
		[]Example{{"Encrypt text", "secret", "ciphertext plus authentication tag"}},
	},
	"sign": {
		"Signs data and verifies signatures using Ed25519, RSA-PSS/SHA-256, or ECDSA/SHA-256.",
		"Use it for signature fixtures, API verification, JWT/JWS-adjacent debugging, and key workflow tests.",
		referenceSets["sign"],
		[]Example{{"Verify", "message", "valid"}},
	},
	"certPrinter": {
		"Parses and prints X.509 certificate details.",
		"Use it to inspect certificate subjects, issuers, validity dates, SANs, key usage, and fingerprints.",
		referenceSets["x509"],
		[]Example{{"Print certificate", "-----BEGIN CERTIFICATE-----...", "Certificate:\n    Data:\n        Version: 3"}},
	},
	"certCloner": {
		"Builds a new certificate based on fields from an existing certificate.",
		"Use it for lab, testing, and proxy workflows where a certificate shape must be reproduced.",
		referenceSets["x509"],
		[]Example{{"Clone certificate", "-----BEGIN CERTIFICATE-----...", "-----BEGIN CERTIFICATE-----\n...\n-----BEGIN PRIVATE KEY-----"}},
	},
}

// UICatalog returns the plugin catalog enriched with descriptions, use cases,
// decode support and references for use in graphical interfaces.
func UICatalog() []UIPluginInfo {
	infos := make([]UIPluginInfo, 0, len(metadata))
	for _, c := range PluginCategories {
		for _, p := range metadata {
			if p.Category != c {
				continue
			}
			copy := copyForPlugin(p.Name, p.Category)
			infos = append(infos, UIPluginInfo{
				Name:        p.Name,
				Label:       PluginLabel(p.Name),
				Category:    p.Category,
				Aliases:     p.Aliases,
				Description: copy.Description,
				UseFor:      copy.UseFor,
				CanDecode:   p.Unprocess != nil,
				References:  copy.References,
				Examples:    copy.Examples,
			})
		}
	}
	return infos
}

// SearchUICatalog returns plugins whose user-facing catalog fields match query.
// Empty queries return the full catalog in normal category order.
func SearchUICatalog(query string) []UIPluginInfo {
	query = strings.ToLower(strings.TrimSpace(query))
	catalog := UICatalog()
	if query == "" {
		return catalog
	}

	var matches []UIPluginInfo
	for _, info := range catalog {
		haystack := strings.ToLower(strings.Join(append([]string{
			info.Name,
			info.Label,
			info.Category,
			CategoryLabel(info.Category),
			info.Description,
			info.UseFor,
		}, info.Aliases...), " "))
		if strings.Contains(haystack, query) {
			matches = append(matches, info)
		}
	}
	return matches
}

func copyForPlugin(name, category string) catalogCopy {
	if c, ok := catalogCopyByName[name]; ok {
		return c
	}
	switch {
	case hasPrefix(name, "sha1"):
		return catalogCopy{Description: "Computes a SHA-1 digest.", UseFor: "Use it for legacy fingerprint checks only. Prefer SHA-256 or SHA-3 for new integrity workflows.", References: referenceSets["sha1"]}
	case hasPrefix(name, "sha2") || hasPrefix(name, "sha224") || hasPrefix(name, "sha256") || hasPrefix(name, "sha384") || hasPrefix(name, "sha512"):
		return catalogCopy{Description: "Computes a SHA-2 digest.", UseFor: "Use it for file fingerprints, integrity checks, and modern non-password hashing.", References: referenceSets["sha2"], Examples: []Example{{"SHA-256 digest", "test", "9f86d081884c7d65..."}}}
	case hasPrefix(name, "sha3"):
		return catalogCopy{Description: "Computes a SHA-3 digest.", UseFor: "Use it for modern integrity checks when SHA-3 compatibility or sponge-based hashing is wanted.", References: referenceSets["sha3"]}
	case hasPrefix(name, "md4") || hasPrefix(name, "md5"):
		return catalogCopy{Description: "Computes an MD-family digest.", UseFor: "Use it for legacy compatibility and quick fingerprints only. It is not suitable for security-sensitive integrity.", References: referenceSets["md"]}
	case hasPrefix(name, "ripemd160"):
		return catalogCopy{Description: "Computes a RIPEMD-160 digest.", UseFor: "Use it for compatibility with systems and fingerprints that specifically require RIPEMD-160.", References: referenceSets["ripemd"]}
	case hasPrefix(name, "blake2"):
		return catalogCopy{Description: "Computes a BLAKE2 digest, with optional keyed hashing for supported variants.", UseFor: "Use it for fast modern hashing, keyed fingerprints, and test vectors.", References: referenceSets["blake2"]}
	case hasPrefix(name, "blake3"):
		return catalogCopy{Description: "Computes a BLAKE3 digest with optional extended output modes.", UseFor: "Use it for very fast modern hashing and variable-length digests.", References: referenceSets["blake3"]}
	case hasPrefix(name, "adler32"):
		return catalogCopy{Description: "Computes an Adler-32 checksum.", UseFor: "Use it for legacy checksum compatibility and accidental-error detection, not cryptographic integrity.", References: referenceSets["adler"]}
	case hasPrefix(name, "crc"):
		return catalogCopy{Description: "Computes a CRC checksum variant.", UseFor: "Use it for protocol checksums, file format compatibility, and accidental-error detection.", References: referenceSets["crc"]}
	case hasPrefix(name, "fnv"):
		return catalogCopy{Description: "Computes an FNV non-cryptographic hash.", UseFor: "Use it for hash-table compatibility, quick bucketing, and legacy fingerprints, not security decisions.", References: referenceSets["fnv"]}
	case category == "hashs":
		return catalogCopy{Description: "Computes a one-way hash or checksum.", UseFor: "Use it when you need a deterministic digest of the input."}
	default:
		return catalogCopy{Description: "Transforms input data.", UseFor: "Use it as part of a deen transform chain when this format or algorithm appears in your data."}
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
