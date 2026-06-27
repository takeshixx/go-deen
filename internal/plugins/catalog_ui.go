package plugins

// Reference points users to background material for a plugin or family.
type Reference struct {
	Label string
	URL   string
}

// UIPluginInfo is catalog metadata shaped for human-facing interfaces.
type UIPluginInfo struct {
	Name        string
	Category    string
	Aliases     []string
	Description string
	UseFor      string
	CanDecode   bool
	References  []Reference
}

type catalogCopy struct {
	Description string
	UseFor      string
	References  []Reference
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
	default:
		return category
	}
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
	"jwt": {
		{"RFC 7519", "https://www.rfc-editor.org/rfc/rfc7519"},
	},
	"jq": {
		{"jq manual", "https://jqlang.github.io/jq/manual/"},
	},
	"x509": {
		{"RFC 5280", "https://www.rfc-editor.org/rfc/rfc5280"},
	},
}

var catalogCopyByName = map[string]catalogCopy{
	"base32": {
		"Encodes binary data as Base32 text and decodes Base32 back to bytes.",
		"Use it when data must travel through uppercase, text-only channels such as tokens, provisioning secrets, or DNS-adjacent workflows.",
		referenceSets["rfc4648"],
	},
	"base64": {
		"Encodes bytes as Base64 text and decodes standard, raw, URL-safe, and raw URL-safe Base64.",
		"Use it for common API payloads, certificates, JWT sections, cookies, and other compact binary-to-text data.",
		referenceSets["rfc4648"],
	},
	"base85": {
		"Encodes bytes using the denser Ascii85/Base85 representation and decodes it again.",
		"Use it when you see compact printable data from PostScript/PDF-style tools or need less overhead than Base64.",
		[]Reference{{"Ascii85 overview", "https://en.wikipedia.org/wiki/Ascii85"}},
	},
	"hex": {
		"Converts bytes to hexadecimal text and parses hexadecimal text back to bytes.",
		"Use it for hashes, binary protocol fragments, keys, signatures, and quick byte-level inspection.",
		nil,
	},
	"url": {
		"Escapes and unescapes URL query text using percent encoding.",
		"Use it to inspect query parameters, callback URLs, webhooks, and payloads copied from browser or proxy traffic.",
		referenceSets["url"],
	},
	"html": {
		"Escapes and unescapes HTML character entities.",
		"Use it when web output contains entities like ampersand escapes or numeric character references.",
		referenceSets["html"],
	},
	"unicode": {
		"Converts text between Unicode escape forms and readable characters.",
		"Use it to inspect escaped strings from JSON, JavaScript, logs, or payloads that hide non-ASCII characters.",
		referenceSets["unicode"],
	},
	"strconv": {
		"Quotes and unquotes strings using Go-style escape sequences.",
		"Use it for debugging string literals, control characters, and copied Go or JSON-like escaped text.",
		[]Reference{{"Go strconv package", "https://pkg.go.dev/strconv"}},
	},
	"pem": {
		"Wraps DER bytes into PEM blocks and unwraps PEM text back to DER bytes.",
		"Use it for certificates, keys, CSRs, and other ASN.1 material exchanged as PEM text.",
		referenceSets["pem"],
	},
	"quoted-printable": {
		"Encodes mostly-readable text as quoted-printable and decodes it again.",
		"Use it for email bodies, MIME parts, and payloads where non-ASCII bytes are represented as equals escapes.",
		referenceSets["quoted-printable"],
	},
	"rot13": {
		"Applies the ROT13 substitution cipher. Running it twice restores the original text.",
		"Use it for simple obfuscation, puzzle text, or legacy examples. It is not encryption.",
		[]Reference{{"ROT13", "https://en.wikipedia.org/wiki/ROT13"}},
	},
	"flate": {
		"Compresses and decompresses raw DEFLATE streams.",
		"Use it for low-level compressed data where there is no gzip or zlib wrapper.",
		referenceSets["deflate"],
	},
	"gzip": {
		"Compresses and decompresses gzip streams.",
		"Use it for HTTP content, log archives, command-line gzip data, and portable compressed blobs.",
		referenceSets["gzip"],
	},
	"zlib": {
		"Compresses and decompresses zlib-wrapped DEFLATE streams.",
		"Use it for PNG-adjacent data, application protocols, or formats that carry zlib streams.",
		referenceSets["zlib"],
	},
	"bzip2": {
		"Compresses and decompresses bzip2 data.",
		"Use it for older archives or data where bzip2's high compression ratio matters more than speed.",
		referenceSets["bzip2"],
	},
	"lzma": {
		"Compresses and decompresses LZMA data.",
		"Use it for 7-Zip/XZ-adjacent workflows and compact archival data.",
		referenceSets["lzma"],
	},
	"lzma2": {
		"Compresses and decompresses LZMA2 data.",
		"Use it for XZ/7-Zip-style data that uses the newer LZMA2 stream format.",
		referenceSets["lzma"],
	},
	"lzw": {
		"Compresses and decompresses LZW data.",
		"Use it for legacy formats and protocol samples that still rely on LZW coding.",
		referenceSets["lzw"],
	},
	"brotli": {
		"Compresses and decompresses Brotli data.",
		"Use it for modern web compression, HTTP responses, and compact static asset payloads.",
		referenceSets["brotli"],
	},
	"zstd": {
		"Compresses and decompresses Zstandard data.",
		"Use it for modern high-speed compression, logs, backups, and large payloads.",
		referenceSets["zstd"],
	},
	"hmac": {
		"Computes keyed message authentication codes with selectable hash algorithms.",
		"Use it to verify webhook signatures, signed API requests, and integrity checks that require a shared secret.",
		referenceSets["hmac"],
	},
	"bcrypt": {
		"Derives password hashes using bcrypt.",
		"Use it for password-hash experiments and verification fixtures. It is intentionally slow and one-way.",
		referenceSets["bcrypt"],
	},
	"scrypt": {
		"Derives memory-hard hashes using scrypt.",
		"Use it for password-derived keys and password-hash testing where memory cost matters.",
		referenceSets["scrypt"],
	},
	"json": {
		"Formats or minifies JSON.",
		"Use it to make API responses readable, normalize JSON before comparing it, or compact JSON for transport.",
		referenceSets["json"],
	},
	"xml": {
		"Formats or minifies XML.",
		"Use it for SOAP, SAML, configuration files, and XML API responses.",
		referenceSets["xml"],
	},
	"json2xml": {
		"Converts JSON to XML and XML back to JSON.",
		"Use it when moving data between services or tooling that expect different structured formats.",
		[]Reference{referenceSets["json"][0], referenceSets["xml"][0]},
	},
	"jwt": {
		"Decodes JWTs into readable header and claim data.",
		"Use it to inspect authentication tokens, claim sets, algorithms, and expiration fields.",
		referenceSets["jwt"],
	},
	"jq": {
		"Runs jq-style queries and filters against JSON input.",
		"Use it to extract fields, reshape responses, and test JSON filters without leaving deen.",
		referenceSets["jq"],
	},
	"certPrinter": {
		"Parses and prints X.509 certificate details.",
		"Use it to inspect certificate subjects, issuers, validity dates, SANs, key usage, and fingerprints.",
		referenceSets["x509"],
	},
	"certCloner": {
		"Builds a new certificate based on fields from an existing certificate.",
		"Use it for lab, testing, and proxy workflows where a certificate shape must be reproduced.",
		referenceSets["x509"],
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
				Category:    p.Category,
				Aliases:     p.Aliases,
				Description: copy.Description,
				UseFor:      copy.UseFor,
				CanDecode:   p.Unprocess != nil,
				References:  copy.References,
			})
		}
	}
	return infos
}

func copyForPlugin(name, category string) catalogCopy {
	if c, ok := catalogCopyByName[name]; ok {
		return c
	}
	switch {
	case hasPrefix(name, "sha1"):
		return catalogCopy{"Computes a SHA-1 digest.", "Use it for legacy fingerprint checks only. Prefer SHA-256 or SHA-3 for new integrity workflows.", referenceSets["sha1"]}
	case hasPrefix(name, "sha2") || hasPrefix(name, "sha224") || hasPrefix(name, "sha256") || hasPrefix(name, "sha384") || hasPrefix(name, "sha512"):
		return catalogCopy{"Computes a SHA-2 digest.", "Use it for file fingerprints, integrity checks, and modern non-password hashing.", referenceSets["sha2"]}
	case hasPrefix(name, "sha3"):
		return catalogCopy{"Computes a SHA-3 digest.", "Use it for modern integrity checks when SHA-3 compatibility or sponge-based hashing is wanted.", referenceSets["sha3"]}
	case hasPrefix(name, "md4") || hasPrefix(name, "md5"):
		return catalogCopy{"Computes an MD-family digest.", "Use it for legacy compatibility and quick fingerprints only. It is not suitable for security-sensitive integrity.", referenceSets["md"]}
	case hasPrefix(name, "ripemd160"):
		return catalogCopy{"Computes a RIPEMD-160 digest.", "Use it for compatibility with systems and fingerprints that specifically require RIPEMD-160.", referenceSets["ripemd"]}
	case hasPrefix(name, "blake2"):
		return catalogCopy{"Computes a BLAKE2 digest, with optional keyed hashing for supported variants.", "Use it for fast modern hashing, keyed fingerprints, and test vectors.", referenceSets["blake2"]}
	case hasPrefix(name, "blake3"):
		return catalogCopy{"Computes a BLAKE3 digest with optional extended output modes.", "Use it for very fast modern hashing and variable-length digests.", referenceSets["blake3"]}
	case hasPrefix(name, "adler32"):
		return catalogCopy{"Computes an Adler-32 checksum.", "Use it for legacy checksum compatibility and accidental-error detection, not cryptographic integrity.", referenceSets["adler"]}
	case hasPrefix(name, "crc"):
		return catalogCopy{"Computes a CRC checksum variant.", "Use it for protocol checksums, file format compatibility, and accidental-error detection.", referenceSets["crc"]}
	case hasPrefix(name, "fnv"):
		return catalogCopy{"Computes an FNV non-cryptographic hash.", "Use it for hash-table compatibility, quick bucketing, and legacy fingerprints, not security decisions.", referenceSets["fnv"]}
	case category == "hashs":
		return catalogCopy{"Computes a one-way hash or checksum.", "Use it when you need a deterministic digest of the input.", nil}
	default:
		return catalogCopy{"Transforms input data.", "Use it as part of a deen transform chain when this format or algorithm appears in your data.", nil}
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
