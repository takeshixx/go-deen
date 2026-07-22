package formatters

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

const urlPartsSchemaVersion = 1

// urlPartsDocument is the stable JSON representation emitted by urlparts.
// The decoded fields are convenient to inspect and edit. Raw contains encoded
// spellings that let the reverse transform preserve distinctions such as %2F
// versus / when the corresponding decoded value has not been changed.
type urlPartsDocument struct {
	Version      int                 `json:"version"`
	Scheme       string              `json:"scheme"`
	Userinfo     *urlPartsUserinfo   `json:"userinfo"`
	Hostname     string              `json:"hostname"`
	Port         string              `json:"port"`
	Opaque       string              `json:"opaque"`
	OmitHost     bool                `json:"omit_host"`
	Path         string              `json:"path"`
	PathSegments []string            `json:"path_segments"`
	Query        []urlQueryParameter `json:"query"`
	ForceQuery   bool                `json:"force_query"`
	Fragment     string              `json:"fragment"`
	Raw          *urlPartsRaw        `json:"raw,omitempty"`
}

type urlPartsUserinfo struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	PasswordSet bool   `json:"password_set"`
}

type urlQueryParameter struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	HasValue bool   `json:"has_value"`
	RawKey   string `json:"raw_key"`
	RawValue string `json:"raw_value"`
}

type urlPartsRaw struct {
	Host     string `json:"host"`
	Path     string `json:"path"`
	Fragment string `json:"fragment"`
}

// NewPluginURLParts creates a reversible URL-to-JSON formatter. It only parses
// local input; it never resolves, opens, or fetches the supplied URL.
func NewPluginURLParts() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "urlparts"
	p.Aliases = []string{".urlparts", "urlparse", ".urlparse"}
	p.Category = "formatters"
	p.Description = "Splits a URL into structured JSON and rebuilds a URL from edited parts."
	p.Process = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		doc, err := parseURLParts(r)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "    ")
		return enc.Encode(doc)
	}
	p.Unprocess = func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		doc, err := decodeURLParts(r)
		if err != nil {
			return err
		}
		rebuilt, err := rebuildURL(doc)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, rebuilt)
		return err
	}
	return p
}

func parseURLParts(r io.Reader) (*urlPartsDocument, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	rawURL := strings.TrimSpace(string(data))
	if rawURL == "" {
		return nil, fmt.Errorf("URL is empty")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	rawPath := u.EscapedPath()
	segments, err := splitURLPath(rawPath)
	if err != nil {
		return nil, fmt.Errorf("decode path: %w", err)
	}
	query, err := splitURLQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	doc := &urlPartsDocument{
		Version:      urlPartsSchemaVersion,
		Scheme:       u.Scheme,
		Hostname:     u.Hostname(),
		Port:         u.Port(),
		Opaque:       u.Opaque,
		OmitHost:     u.OmitHost,
		Path:         u.Path,
		PathSegments: segments,
		Query:        query,
		ForceQuery:   u.ForceQuery,
		Fragment:     u.Fragment,
		Raw: &urlPartsRaw{
			Host:     u.Host,
			Path:     rawPath,
			Fragment: u.EscapedFragment(),
		},
	}
	if u.User != nil {
		password, passwordSet := u.User.Password()
		doc.Userinfo = &urlPartsUserinfo{
			Username:    u.User.Username(),
			Password:    password,
			PasswordSet: passwordSet,
		}
	}
	return doc, nil
}

func splitURLPath(rawPath string) ([]string, error) {
	if rawPath == "" || rawPath == "/" {
		return []string{}, nil
	}
	path := strings.TrimPrefix(rawPath, "/")
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	if path == "" {
		return []string{}, nil
	}
	rawSegments := strings.Split(path, "/")
	segments := make([]string, len(rawSegments))
	for i, segment := range rawSegments {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return nil, fmt.Errorf("segment %d: %w", i+1, err)
		}
		segments[i] = decoded
	}
	return segments, nil
}

func splitURLQuery(rawQuery string) ([]urlQueryParameter, error) {
	if rawQuery == "" {
		return []urlQueryParameter{}, nil
	}
	rawParameters := strings.Split(rawQuery, "&")
	parameters := make([]urlQueryParameter, 0, len(rawParameters))
	for i, parameter := range rawParameters {
		rawKey, rawValue, hasValue := strings.Cut(parameter, "=")
		key, err := url.QueryUnescape(rawKey)
		if err != nil {
			return nil, fmt.Errorf("decode query parameter %d key: %w", i+1, err)
		}
		value := ""
		if hasValue {
			value, err = url.QueryUnescape(rawValue)
			if err != nil {
				return nil, fmt.Errorf("decode query parameter %d value: %w", i+1, err)
			}
		}
		parameters = append(parameters, urlQueryParameter{
			Key:      key,
			Value:    value,
			HasValue: hasValue,
			RawKey:   rawKey,
			RawValue: rawValue,
		})
	}
	return parameters, nil
}

func decodeURLParts(r io.Reader) (*urlPartsDocument, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var doc urlPartsDocument
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode URL parts JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("decode URL parts JSON: multiple JSON values")
		}
		return nil, fmt.Errorf("decode URL parts JSON: %w", err)
	}
	if doc.Version != urlPartsSchemaVersion {
		return nil, fmt.Errorf("unsupported URL parts schema version %d (want %d)", doc.Version, urlPartsSchemaVersion)
	}
	return &doc, nil
}

func rebuildURL(doc *urlPartsDocument) (string, error) {
	if err := validateURLParts(doc); err != nil {
		return "", err
	}
	rawQuery, err := joinURLQuery(doc.Query)
	if err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme:     doc.Scheme,
		Opaque:     doc.Opaque,
		OmitHost:   doc.OmitHost,
		RawQuery:   rawQuery,
		ForceQuery: doc.ForceQuery,
		Fragment:   doc.Fragment,
	}
	if doc.Userinfo != nil {
		if doc.Userinfo.PasswordSet {
			u.User = url.UserPassword(doc.Userinfo.Username, doc.Userinfo.Password)
		} else {
			u.User = url.User(doc.Userinfo.Username)
		}
	}
	u.Host = rebuiltHost(doc)
	setRebuiltPath(u, doc)
	if doc.Raw != nil {
		if decoded, err := url.PathUnescape(doc.Raw.Fragment); err == nil && decoded == doc.Fragment {
			u.RawFragment = doc.Raw.Fragment
		}
	}
	return u.String(), nil
}

func validateURLParts(doc *urlPartsDocument) error {
	if doc.Scheme != "" {
		for i, r := range doc.Scheme {
			valid := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
			if i > 0 {
				valid = valid || r >= '0' && r <= '9' || r == '+' || r == '-' || r == '.'
			}
			if !valid {
				return fmt.Errorf("invalid URL scheme %q", doc.Scheme)
			}
		}
	}
	if doc.Port != "" {
		if doc.Hostname == "" {
			return fmt.Errorf("URL port requires a hostname")
		}
		port, err := strconv.Atoi(doc.Port)
		if err != nil || port < 0 || port > 65535 {
			return fmt.Errorf("invalid URL port %q", doc.Port)
		}
	}
	if doc.Opaque != "" && (doc.Hostname != "" || doc.Port != "" || doc.Path != "" || len(doc.PathSegments) > 0) {
		return fmt.Errorf("opaque URL data cannot be combined with host or path fields")
	}
	return nil
}

func rebuiltHost(doc *urlPartsDocument) string {
	if doc.Raw != nil {
		raw := &url.URL{Host: doc.Raw.Host}
		if raw.Hostname() == doc.Hostname && raw.Port() == doc.Port {
			return doc.Raw.Host
		}
	}
	if doc.Port != "" {
		return net.JoinHostPort(doc.Hostname, doc.Port)
	}
	if strings.Contains(doc.Hostname, ":") && !strings.HasPrefix(doc.Hostname, "[") {
		return "[" + doc.Hostname + "]"
	}
	return doc.Hostname
}

func setRebuiltPath(u *url.URL, doc *urlPartsDocument) {
	if doc.Opaque != "" {
		return
	}
	if doc.Raw != nil {
		originalPath, pathErr := url.PathUnescape(doc.Raw.Path)
		originalSegments, segmentErr := splitURLPath(doc.Raw.Path)
		switch {
		case pathErr == nil && segmentErr == nil && originalPath == doc.Path && slices.Equal(originalSegments, doc.PathSegments):
			u.Path = doc.Path
			u.RawPath = doc.Raw.Path
			return
		case pathErr == nil && segmentErr == nil && originalPath == doc.Path:
			setPathFromSegments(u, doc.Path, doc.PathSegments)
			return
		}
	}
	if doc.Path == "" && len(doc.PathSegments) > 0 {
		setPathFromSegments(u, doc.Path, doc.PathSegments)
		return
	}
	u.Path = doc.Path
}

func setPathFromSegments(u *url.URL, originalPath string, segments []string) {
	rawSegments := make([]string, len(segments))
	for i, segment := range segments {
		rawSegments[i] = url.PathEscape(segment)
	}
	rawPath := strings.Join(rawSegments, "/")
	if strings.HasPrefix(originalPath, "/") {
		rawPath = "/" + rawPath
	}
	if originalPath != "/" && strings.HasSuffix(originalPath, "/") {
		rawPath += "/"
	}
	if originalPath == "/" && len(segments) == 0 {
		rawPath = "/"
	}
	path, _ := url.PathUnescape(rawPath)
	u.Path = path
	u.RawPath = rawPath
}

func joinURLQuery(parameters []urlQueryParameter) (string, error) {
	rawParameters := make([]string, len(parameters))
	for i, parameter := range parameters {
		if !parameter.HasValue && parameter.Value != "" {
			return "", fmt.Errorf("query parameter %d has a value but has_value is false", i+1)
		}
		rawKey := encodedQueryPart(parameter.RawKey, parameter.Key)
		if !parameter.HasValue {
			rawParameters[i] = rawKey
			continue
		}
		rawValue := encodedQueryPart(parameter.RawValue, parameter.Value)
		rawParameters[i] = rawKey + "=" + rawValue
	}
	return strings.Join(rawParameters, "&"), nil
}

func encodedQueryPart(raw, decoded string) string {
	if value, err := url.QueryUnescape(raw); err == nil && value == decoded {
		return raw
	}
	return url.QueryEscape(decoded)
}
