package formatters

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestURLPartsProcessExposesStructuredComponents(t *testing.T) {
	p := NewPluginURLParts()
	input := "https://login-update.example.invalid:8443/account/verify.php?recipient=alice%40corp.example&campaign=Q3&campaign=retry#continue"
	out := runFormat(t, p.Process, p.RegisterFlags, []byte(input))

	var doc urlPartsDocument
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != urlPartsSchemaVersion || doc.Scheme != "https" {
		t.Fatalf("unexpected schema or scheme: %#v", doc)
	}
	if doc.Hostname != "login-update.example.invalid" || doc.Port != "8443" {
		t.Fatalf("unexpected authority: hostname=%q port=%q", doc.Hostname, doc.Port)
	}
	wantSegments := []string{"account", "verify.php"}
	if len(doc.PathSegments) != len(wantSegments) {
		t.Fatalf("path segments = %#v, want %#v", doc.PathSegments, wantSegments)
	}
	for i := range wantSegments {
		if doc.PathSegments[i] != wantSegments[i] {
			t.Fatalf("path segment %d = %q, want %q", i, doc.PathSegments[i], wantSegments[i])
		}
	}
	if len(doc.Query) != 3 || doc.Query[0].Key != "recipient" || doc.Query[0].Value != "alice@corp.example" {
		t.Fatalf("unexpected query: %#v", doc.Query)
	}
	if doc.Query[1].Key != "campaign" || doc.Query[2].Key != "campaign" {
		t.Fatalf("duplicate query keys were not retained: %#v", doc.Query)
	}
	if doc.Fragment != "continue" {
		t.Fatalf("fragment = %q, want continue", doc.Fragment)
	}
}

func TestURLPartsRoundTripPreservesRawEncodingAndQueryShape(t *testing.T) {
	p := NewPluginURLParts()
	input := "https://user:p%40ss@[2001:db8::1]:8443/a%2Fb//c/?first=one+two&dup=1&empty=&flag&dup=%32&&encoded=%2f#frag%2Fment"
	jsonData := runFormat(t, p.Process, p.RegisterFlags, []byte(input))
	got := runFormat(t, p.Unprocess, p.RegisterFlags, jsonData)
	if string(got) != input {
		t.Fatalf("round trip = %q, want %q", got, input)
	}
}

func TestURLPartsEditQueryValueReencodesOnlyChangedValue(t *testing.T) {
	p := NewPluginURLParts()
	input := "https://example.invalid/verify?recipient=alice%40corp.example&space=one%20two&flag"
	jsonData := runFormat(t, p.Process, p.RegisterFlags, []byte(input))
	var doc urlPartsDocument
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		t.Fatal(err)
	}
	doc.Query[0].Value = "bob@corp.example"
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	got := runFormat(t, p.Unprocess, p.RegisterFlags, edited)
	want := "https://example.invalid/verify?recipient=bob%40corp.example&space=one%20two&flag"
	if string(got) != want {
		t.Fatalf("rebuilt URL = %q, want %q", got, want)
	}
}

func TestURLPartsEditPathSegmentsPreservesSegmentBoundaries(t *testing.T) {
	p := NewPluginURLParts()
	jsonData := runFormat(t, p.Process, p.RegisterFlags, []byte("https://example.invalid/a%2Fb/c/"))
	var doc urlPartsDocument
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		t.Fatal(err)
	}
	doc.PathSegments[1] = "d/e"
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	got := runFormat(t, p.Unprocess, p.RegisterFlags, edited)
	if want := "https://example.invalid/a%2Fb/d%2Fe/"; string(got) != want {
		t.Fatalf("rebuilt URL = %q, want %q", got, want)
	}
}

func TestURLPartsEditedPathTakesPrecedenceOverDerivedSegments(t *testing.T) {
	p := NewPluginURLParts()
	jsonData := runFormat(t, p.Process, p.RegisterFlags, []byte("https://example.invalid/original/path"))
	var doc urlPartsDocument
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		t.Fatal(err)
	}
	doc.Path = "/changed path"
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	got := runFormat(t, p.Unprocess, p.RegisterFlags, edited)
	if want := "https://example.invalid/changed%20path"; string(got) != want {
		t.Fatalf("rebuilt URL = %q, want %q", got, want)
	}
}

func TestURLPartsRoundTripsOpaqueURL(t *testing.T) {
	p := NewPluginURLParts()
	input := "mailto:security@example.org?subject=Suspicious+URL#details"
	jsonData := runFormat(t, p.Process, p.RegisterFlags, []byte(input))
	got := runFormat(t, p.Unprocess, p.RegisterFlags, jsonData)
	if string(got) != input {
		t.Fatalf("round trip = %q, want %q", got, input)
	}
}

func TestURLPartsRejectsInvalidInput(t *testing.T) {
	p := NewPluginURLParts()
	tests := []struct {
		name    string
		input   string
		process bool
		want    string
	}{
		{name: "empty URL", input: " \n", process: true, want: "URL is empty"},
		{name: "bad query escape", input: "https://example.invalid/?value=%zz", process: true, want: "decode query parameter"},
		{name: "bad JSON", input: `{`, want: "decode URL parts JSON"},
		{name: "wrong schema", input: `{"version":2}`, want: "unsupported URL parts schema version"},
		{name: "unknown field", input: `{"version":1,"extra":true}`, want: "unknown field"},
		{name: "value without equals", input: `{"version":1,"query":[{"key":"flag","value":"unexpected","has_value":false}]}`, want: "has a value"},
		{name: "bad port", input: `{"version":1,"hostname":"example.invalid","port":"99999"}`, want: "invalid URL port"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.process {
				_, err = tryFormat(p.Process, p.RegisterFlags, []byte(tc.input))
			} else {
				_, err = tryFormat(p.Unprocess, p.RegisterFlags, []byte(tc.input))
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}
