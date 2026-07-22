package formatters

import (
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"unicode"
)

// URLPartsRedirectMaxDepth bounds local recursive redirect inspection.
const URLPartsRedirectMaxDepth = 4

// URLPartsAnalysis contains derived, local-only observations about a URL.
// It is informational and is never used to rebuild the URL.
type URLPartsAnalysis struct {
	Indicators         []URLPartsIndicator         `json:"indicators"`
	NestedURLs         []URLPartsNestedURL         `json:"nested_urls"`
	RedirectChains     []URLPartsRedirectNode      `json:"redirect_chains"`
	TrackingParameters []URLPartsTrackingParameter `json:"tracking_parameters"`
	DefangedURL        string                      `json:"defanged_url"`
}

// URLPartsIndicator describes an observable URL characteristic without making
// a malicious/benign verdict.
type URLPartsIndicator struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// URLPartsNestedURL identifies an HTTP(S) URL found in a decoded query value.
// ParameterIndex is one-based so it maps directly to the ordered query array
// shown by the CLI and graphical editors.
type URLPartsNestedURL struct {
	ParameterIndex int    `json:"parameter_index"`
	Key            string `json:"key"`
	URL            string `json:"url"`
	Scheme         string `json:"scheme"`
	Hostname       string `json:"hostname"`
}

// URLPartsRedirectNode represents one locally decoded HTTP(S) URL in a
// recursive redirect tree. Children are URLs found in this node's query values.
type URLPartsRedirectNode struct {
	Depth          int                    `json:"depth"`
	ParameterIndex int                    `json:"parameter_index"`
	Key            string                 `json:"key"`
	URL            string                 `json:"url"`
	DefangedURL    string                 `json:"defanged_url"`
	Scheme         string                 `json:"scheme"`
	Hostname       string                 `json:"hostname"`
	Cycle          bool                   `json:"cycle"`
	Truncated      bool                   `json:"truncated"`
	Children       []URLPartsRedirectNode `json:"children"`
}

// URLPartsTrackingParameter identifies a commonly used analytics or campaign
// parameter. ParameterIndex is one-based.
type URLPartsTrackingParameter struct {
	ParameterIndex int    `json:"parameter_index"`
	Key            string `json:"key"`
}

// AnalyzeURLParts derives indicators, first-hop nested URLs, bounded redirect
// chains, tracking parameters, and a defanged representation without resolving
// or fetching the URL.
func AnalyzeURLParts(doc *URLPartsDocument) (*URLPartsAnalysis, error) {
	rebuilt, err := RebuildURLParts(doc)
	if err != nil {
		return nil, err
	}
	defanged, err := DefangURLParts(doc)
	if err != nil {
		return nil, err
	}
	analysis := &URLPartsAnalysis{
		Indicators:         []URLPartsIndicator{},
		NestedURLs:         []URLPartsNestedURL{},
		RedirectChains:     []URLPartsRedirectNode{},
		TrackingParameters: []URLPartsTrackingParameter{},
		DefangedURL:        defanged,
	}

	if doc.Userinfo != nil {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "embedded_credentials",
			Severity: "warning",
			Message:  "The URL contains user-information before the hostname.",
		})
	}
	if isIPLiteral(doc.Hostname) {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "ip_literal_host",
			Severity: "warning",
			Message:  "The hostname is an IP address rather than a domain name.",
		})
	}
	if isInternationalizedHostname(doc.Hostname) {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "internationalized_hostname",
			Severity: "warning",
			Message:  "The hostname contains Unicode or Punycode labels; inspect its spelling carefully.",
		})
	}
	if strings.EqualFold(doc.Scheme, "http") {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "cleartext_http",
			Severity: "warning",
			Message:  "The URL uses HTTP instead of HTTPS.",
		})
	}
	if (strings.EqualFold(doc.Scheme, "http") || strings.EqualFold(doc.Scheme, "https")) && doc.Hostname == "" {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "missing_hostname",
			Severity: "warning",
			Message:  "The HTTP(S) URL has no hostname.",
		})
	}
	switch strings.ToLower(doc.Scheme) {
	case "javascript", "data", "file":
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "active_or_local_scheme",
			Severity: "warning",
			Message:  fmt.Sprintf("The URL uses the %s scheme, which may execute content or access local resources when opened.", doc.Scheme),
		})
	}
	if isNonDefaultPort(doc.Scheme, doc.Port) {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "non_default_port",
			Severity: "info",
			Message:  fmt.Sprintf("The URL uses explicit non-default port %s.", doc.Port),
		})
	}

	for i, parameter := range doc.Query {
		if isTrackingParameter(parameter.Key) {
			analysis.TrackingParameters = append(analysis.TrackingParameters, URLPartsTrackingParameter{
				ParameterIndex: i + 1,
				Key:            parameter.Key,
			})
		}
	}
	visited := map[string]bool{rebuilt: true}
	analysis.RedirectChains = buildURLRedirectNodes(doc.Query, 1, visited)
	for _, node := range analysis.RedirectChains {
		analysis.NestedURLs = append(analysis.NestedURLs, URLPartsNestedURL{
			ParameterIndex: node.ParameterIndex,
			Key:            node.Key,
			URL:            node.URL,
			Scheme:         node.Scheme,
			Hostname:       node.Hostname,
		})
	}
	redirectCount := countURLRedirectNodes(analysis.RedirectChains)
	if redirectCount > 0 {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "nested_url",
			Severity: "info",
			Message:  fmt.Sprintf("Found %d nested HTTP(S) URL(s) in query values.", redirectCount),
		})
	}
	if len(analysis.TrackingParameters) > 0 {
		analysis.Indicators = append(analysis.Indicators, URLPartsIndicator{
			Code:     "tracking_parameters",
			Severity: "info",
			Message:  fmt.Sprintf("Found %d common tracking parameter occurrence(s).", len(analysis.TrackingParameters)),
		})
	}
	return analysis, nil
}

// DefangURLParts defangs both the outer URL and recognized nested HTTP(S)
// query values so copied text cannot retain a live nested URL substring.
func DefangURLParts(doc *URLPartsDocument) (string, error) {
	rebuilt, err := RebuildURLParts(doc)
	if err != nil {
		return "", err
	}
	copyDoc := *doc
	copyDoc.Query = append([]URLQueryParameter(nil), doc.Query...)
	changed := false
	for i := range copyDoc.Query {
		parameter := &copyDoc.Query[i]
		if !parameter.HasValue {
			continue
		}
		nested, ok := findNestedURL(parameter.Value)
		if !ok {
			continue
		}
		defanged, err := DefangURL(nested.String())
		if err != nil {
			continue
		}
		parameter.Value = defanged
		parameter.RawValue = ""
		changed = true
	}
	if changed {
		rebuilt, err = RebuildURLParts(&copyDoc)
		if err != nil {
			return "", err
		}
	}
	return DefangURL(rebuilt)
}

func buildURLRedirectNodes(parameters []URLQueryParameter, depth int, visited map[string]bool) []URLPartsRedirectNode {
	nodes := []URLPartsRedirectNode{}
	for i, parameter := range parameters {
		if !parameter.HasValue {
			continue
		}
		nested, ok := findNestedURL(parameter.Value)
		if !ok {
			continue
		}
		nestedURL := nested.String()
		defanged := defangNestedURL(nestedURL)
		node := URLPartsRedirectNode{
			Depth:          depth,
			ParameterIndex: i + 1,
			Key:            parameter.Key,
			URL:            nestedURL,
			DefangedURL:    defanged,
			Scheme:         nested.Scheme,
			Hostname:       nested.Hostname(),
			Children:       []URLPartsRedirectNode{},
		}
		if visited[nestedURL] {
			node.Cycle = true
			nodes = append(nodes, node)
			continue
		}
		query, err := splitURLQuery(nested.RawQuery)
		if err != nil {
			nodes = append(nodes, node)
			continue
		}
		if depth >= URLPartsRedirectMaxDepth {
			node.Truncated = hasNestedURL(query)
			nodes = append(nodes, node)
			continue
		}
		nextVisited := make(map[string]bool, len(visited)+1)
		for value := range visited {
			nextVisited[value] = true
		}
		nextVisited[nestedURL] = true
		node.Children = buildURLRedirectNodes(query, depth+1, nextVisited)
		nodes = append(nodes, node)
	}
	return nodes
}

func defangNestedURL(rawURL string) string {
	doc, err := parseURLParts(strings.NewReader(rawURL))
	if err == nil {
		if defanged, err := DefangURLParts(doc); err == nil {
			return defanged
		}
	}
	defanged, _ := DefangURL(rawURL)
	return defanged
}

func hasNestedURL(parameters []URLQueryParameter) bool {
	for _, parameter := range parameters {
		if parameter.HasValue {
			if _, ok := findNestedURL(parameter.Value); ok {
				return true
			}
		}
	}
	return false
}

func countURLRedirectNodes(nodes []URLPartsRedirectNode) int {
	count := 0
	for _, node := range nodes {
		count += 1 + countURLRedirectNodes(node.Children)
	}
	return count
}

// DefangURL makes URLs safer to paste into tickets and chat. HTTP(S) becomes
// hxxp(s), other schemes have their colon neutralized, and authority dots
// become [.].
func DefangURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse URL for defanging: %w", err)
	}
	defanged := rawURL
	if u.Host != "" {
		if marker := strings.Index(defanged, "//"); marker >= 0 {
			authorityStart := marker + 2
			authorityEnd := len(defanged)
			if offset := strings.IndexAny(defanged[authorityStart:], "/?#"); offset >= 0 {
				authorityEnd = authorityStart + offset
			}
			authority := defanged[authorityStart:authorityEnd]
			userinfoEnd := strings.LastIndex(authority, "@") + 1
			hostport := authority[userinfoEnd:]
			hostport = strings.ReplaceAll(hostport, ".", "[.]")
			defanged = defanged[:authorityStart+userinfoEnd] + hostport + defanged[authorityEnd:]
		}
	}
	if u.Scheme != "" {
		if colon := strings.Index(defanged, ":"); colon >= 0 {
			replacement := defanged[:colon] + "[:]"
			switch {
			case strings.EqualFold(u.Scheme, "http"):
				replacement = "hxxp:"
			case strings.EqualFold(u.Scheme, "https"):
				replacement = "hxxps:"
			}
			defanged = replacement + defanged[colon+1:]
		}
	}
	return defanged, nil
}

// RemoveURLTrackingParameter removes one detected tracking parameter by its
// one-based query position. It returns false when the position is invalid or
// no longer refers to a recognized tracking key.
func RemoveURLTrackingParameter(doc *URLPartsDocument, parameterIndex int) bool {
	if doc == nil || parameterIndex < 1 || parameterIndex > len(doc.Query) {
		return false
	}
	index := parameterIndex - 1
	if !isTrackingParameter(doc.Query[index].Key) {
		return false
	}
	doc.Query = slices.Delete(doc.Query, index, index+1)
	refreshURLPartsAnalysis(doc)
	return true
}

// RemoveAllURLTrackingParameters removes every currently recognized tracking
// parameter while retaining the order and raw spellings of all other entries.
func RemoveAllURLTrackingParameters(doc *URLPartsDocument) int {
	if doc == nil || len(doc.Query) == 0 {
		return 0
	}
	before := len(doc.Query)
	doc.Query = slices.DeleteFunc(doc.Query, func(parameter URLQueryParameter) bool {
		return isTrackingParameter(parameter.Key)
	})
	removed := before - len(doc.Query)
	refreshURLPartsAnalysis(doc)
	return removed
}

func refreshURLPartsAnalysis(doc *URLPartsDocument) {
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		doc.Analysis = nil
		return
	}
	doc.Analysis = analysis
}

func isIPLiteral(hostname string) bool {
	hostname, _, _ = strings.Cut(hostname, "%")
	return net.ParseIP(hostname) != nil
}

func isInternationalizedHostname(hostname string) bool {
	for _, label := range strings.Split(hostname, ".") {
		if strings.HasPrefix(strings.ToLower(label), "xn--") {
			return true
		}
	}
	return strings.IndexFunc(hostname, func(r rune) bool { return r > unicode.MaxASCII }) >= 0
}

func isNonDefaultPort(scheme, port string) bool {
	if port == "" {
		return false
	}
	switch strings.ToLower(scheme) {
	case "http":
		return port != "80"
	case "https":
		return port != "443"
	default:
		return true
	}
}

func findNestedURL(value string) (*url.URL, bool) {
	candidate := strings.TrimSpace(value)
	for range 3 {
		u, err := url.Parse(candidate)
		if err == nil && u.Hostname() != "" && (strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) {
			return u, true
		}
		if !strings.Contains(candidate, "%") {
			break
		}
		decoded, err := url.QueryUnescape(candidate)
		if err != nil || decoded == candidate {
			break
		}
		candidate = decoded
	}
	return nil, false
}

func isTrackingParameter(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, prefix := range []string{"utm_", "pk_", "mtm_", "matomo_", "hsa_"} {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	switch key {
	case "_ga", "_gl", "dclid", "fbclid", "gclid", "igshid", "li_fat_id", "mc_cid", "mc_eid", "mkt_tok", "msclkid", "ttclid", "twclid", "vero_id":
		return true
	default:
		return false
	}
}
