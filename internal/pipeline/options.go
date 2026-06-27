package pipeline

import (
	"flag"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/internal/plugins"
)

// Option describes a single configurable plugin flag for UI rendering.
type Option struct {
	Name    string
	Label   string
	Usage   string
	Default string
	IsBool  bool
	Kind    string
	Choices []string
	Secret  bool
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
			Name:    f.Name,
			Label:   optionLabel(p.Name, f.Name),
			Usage:   f.Usage,
			Default: f.DefValue,
			IsBool:  isBool,
			Kind:    kind,
			Choices: optionChoices(p.Name, f.Name),
			Secret:  isSecretOption(f.Name),
		})
	})
	return opts
}

func optionLabel(plugin, name string) string {
	switch plugin + ":" + name {
	case "jwt:r":
		return "Recreate token, keep signature"
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
	default:
		return name
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
