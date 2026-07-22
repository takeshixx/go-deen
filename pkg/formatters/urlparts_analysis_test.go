package formatters

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestURLPartsAnalysisFindsIndicatorsNestedURLsAndTracking(t *testing.T) {
	p := NewPluginURLParts()
	input := "http://user:secret@192.0.2.10:8080/login?utm_source=mail&redirect=https%253A%252F%252Fportal.example.org%252Fsignin&flag"
	out := runFormat(t, p.Process, p.RegisterFlags, []byte(input))

	var doc URLPartsDocument
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != 2 || doc.Analysis == nil {
		t.Fatalf("version=%d analysis=%#v", doc.Version, doc.Analysis)
	}
	gotCodes := make([]string, len(doc.Analysis.Indicators))
	for i, indicator := range doc.Analysis.Indicators {
		gotCodes[i] = indicator.Code
	}
	wantCodes := []string{"embedded_credentials", "ip_literal_host", "cleartext_http", "non_default_port", "nested_url", "tracking_parameters"}
	if !slices.Equal(gotCodes, wantCodes) {
		t.Fatalf("indicator codes = %#v, want %#v", gotCodes, wantCodes)
	}
	if len(doc.Analysis.NestedURLs) != 1 {
		t.Fatalf("nested URLs = %#v", doc.Analysis.NestedURLs)
	}
	nested := doc.Analysis.NestedURLs[0]
	if nested.ParameterIndex != 2 || nested.Key != "redirect" || nested.URL != "https://portal.example.org/signin" || nested.Hostname != "portal.example.org" {
		t.Fatalf("nested URL = %#v", nested)
	}
	if len(doc.Analysis.TrackingParameters) != 1 || doc.Analysis.TrackingParameters[0].ParameterIndex != 1 || doc.Analysis.TrackingParameters[0].Key != "utm_source" {
		t.Fatalf("tracking parameters = %#v", doc.Analysis.TrackingParameters)
	}
	wantDefanged := "hxxp://user:secret@192[.]0[.]2[.]10:8080/login?utm_source=mail&redirect=https%253A%252F%252Fportal.example.org%252Fsignin&flag"
	if doc.Analysis.DefangedURL != wantDefanged {
		t.Fatalf("defanged URL = %q, want %q", doc.Analysis.DefangedURL, wantDefanged)
	}
}

func TestURLPartsAnalysisFlagsInternationalizedHostname(t *testing.T) {
	doc := &URLPartsDocument{Version: URLPartsSchemaVersion, Scheme: "https", Hostname: "xn--pple-43d.example"}
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Indicators) != 1 || analysis.Indicators[0].Code != "internationalized_hostname" {
		t.Fatalf("indicators = %#v", analysis.Indicators)
	}
	if analysis.DefangedURL != "hxxps://xn--pple-43d[.]example" {
		t.Fatalf("defanged URL = %q", analysis.DefangedURL)
	}
}

func TestURLPartsAnalysisNeutralizesActiveScheme(t *testing.T) {
	doc := &URLPartsDocument{Version: URLPartsSchemaVersion, Scheme: "javascript", Opaque: "alert(1)"}
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Indicators) != 1 || analysis.Indicators[0].Code != "active_or_local_scheme" {
		t.Fatalf("indicators = %#v", analysis.Indicators)
	}
	if analysis.DefangedURL != "javascript[:]alert(1)" {
		t.Fatalf("defanged URL = %q", analysis.DefangedURL)
	}
}

func TestURLPartsAnalysisFlagsHTTPWithoutHostname(t *testing.T) {
	doc := &URLPartsDocument{Version: URLPartsSchemaVersion, Scheme: "https", Path: "relative"}
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Indicators) != 1 || analysis.Indicators[0].Code != "missing_hostname" {
		t.Fatalf("indicators = %#v", analysis.Indicators)
	}
}

func TestURLPartsDecodeAcceptsVersionOneAndEncodeUpgradesIt(t *testing.T) {
	doc, err := DecodeURLPartsJSON(strings.NewReader(`{"version":1,"scheme":"https","hostname":"example.invalid"}`))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Analysis == nil {
		t.Fatal("version 1 document was not analyzed")
	}
	var out bytes.Buffer
	if err := EncodeURLPartsJSON(&out, doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != URLPartsSchemaVersion {
		t.Fatalf("encoded document version = %d", doc.Version)
	}
	if !strings.Contains(out.String(), `"version": 2`) || !strings.Contains(out.String(), `"analysis"`) {
		t.Fatalf("upgraded JSON = %s", out.String())
	}
}

func TestURLPartsAnalysisRefreshesAfterEdit(t *testing.T) {
	doc := &URLPartsDocument{
		Version:  URLPartsSchemaVersion,
		Scheme:   "https",
		Hostname: "example.invalid",
		Query: []URLQueryParameter{{
			Key:      "redirect",
			Value:    "https://first.invalid/",
			HasValue: true,
		}},
	}
	var first bytes.Buffer
	if err := EncodeURLPartsJSON(&first, doc); err != nil {
		t.Fatal(err)
	}
	doc.Query[0].Value = "https://second.invalid/path"
	var second bytes.Buffer
	if err := EncodeURLPartsJSON(&second, doc); err != nil {
		t.Fatal(err)
	}
	if doc.Analysis == nil || len(doc.Analysis.NestedURLs) != 1 || doc.Analysis.NestedURLs[0].Hostname != "second.invalid" {
		t.Fatalf("refreshed analysis = %#v", doc.Analysis)
	}
	if strings.Contains(second.String(), "first.invalid") {
		t.Fatalf("stale analysis remained in JSON: %s", second.String())
	}
}
