package formatters

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginJWK creates a formatter for JSON Web Keys and JSON Web Key Sets.
func NewPluginJWK() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "jwk"
	p.Aliases = []string{".jwk", "jwks", ".jwks"}
	p.Category = "formatters"
	p.Description = "Format, compact and inspect JSON Web Keys (JWK) and JSON Web Key Sets (JWKS)."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.Bool("public", false, "emit public keys when possible")
		flags.Bool("thumbprint", false, "emit RFC 7638 SHA-256 thumbprints instead of keys")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		out, err := formatJWK(data, helpers.IsBoolFlag(flags, "public"), helpers.IsBoolFlag(flags, "thumbprint"), true)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		out, err := formatJWK(data, helpers.IsBoolFlag(flags, "public"), helpers.IsBoolFlag(flags, "thumbprint"), false)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}
	return p
}

func formatJWK(data []byte, publicOnly, thumbprint, pretty bool) ([]byte, error) {
	if isJWKS(data) {
		var set jose.JSONWebKeySet
		if err := json.Unmarshal(data, &set); err != nil {
			return nil, err
		}
		if thumbprint {
			out, err := thumbprintSet(set, publicOnly)
			if err != nil {
				return nil, err
			}
			return marshalJSON(out, pretty)
		}
		if publicOnly {
			for i := range set.Keys {
				set.Keys[i] = publicJWK(set.Keys[i])
			}
		}
		return marshalJSON(set, pretty)
	}

	var key jose.JSONWebKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, err
	}
	if thumbprint {
		out, err := thumbprintKey(key, publicOnly)
		if err != nil {
			return nil, err
		}
		return marshalJSON(out, pretty)
	}
	if publicOnly {
		key = publicJWK(key)
	}
	return marshalJSON(key, pretty)
}

func isJWKS(data []byte) bool {
	var probe struct {
		Keys []json.RawMessage `json:"keys"`
	}
	return json.Unmarshal(data, &probe) == nil && probe.Keys != nil
}

func publicJWK(key jose.JSONWebKey) jose.JSONWebKey {
	pub := key.Public()
	if pub.Valid() {
		return pub
	}
	return key
}

type jwkThumbprint struct {
	KeyID              string `json:"kid,omitempty"`
	ThumbprintSHA256   string `json:"thumbprint_sha256"`
	PublicThumbprinted bool   `json:"public_thumbprinted,omitempty"`
}

type jwksThumbprints struct {
	Keys []jwkThumbprint `json:"keys"`
}

func thumbprintSet(set jose.JSONWebKeySet, publicOnly bool) (jwksThumbprints, error) {
	out := jwksThumbprints{Keys: make([]jwkThumbprint, 0, len(set.Keys))}
	for _, key := range set.Keys {
		tp, err := thumbprintKey(key, publicOnly)
		if err != nil {
			return out, err
		}
		out.Keys = append(out.Keys, tp)
	}
	return out, nil
}

func thumbprintKey(key jose.JSONWebKey, publicOnly bool) (jwkThumbprint, error) {
	if publicOnly {
		key = publicJWK(key)
	}
	tp, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		return jwkThumbprint{}, err
	}
	return jwkThumbprint{
		KeyID:              key.KeyID,
		ThumbprintSHA256:   base64.RawURLEncoding.EncodeToString(tp),
		PublicThumbprinted: publicOnly,
	}, nil
}

func marshalJSON(v any, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(v, "", "  ")
	}
	out, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JWK JSON: %w", err)
	}
	return out, nil
}
