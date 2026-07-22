package formatters

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestURLPartsAnalysisBuildsRecursiveRedirectTree(t *testing.T) {
	inner := "https://third.invalid/final"
	second := "https://second.invalid/next?target=" + urlQueryEscapeForTest(inner)
	first := "https://first.invalid/start?redirect=" + urlQueryEscapeForTest(second)
	doc := mustParseURLPartsForTest(t, "https://root.invalid/?next="+urlQueryEscapeForTest(first))
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.RedirectChains) != 1 || len(analysis.RedirectChains[0].Children) != 1 || len(analysis.RedirectChains[0].Children[0].Children) != 1 {
		t.Fatalf("redirect chains = %#v", analysis.RedirectChains)
	}
	third := analysis.RedirectChains[0].Children[0].Children[0]
	if third.Depth != 3 || third.Hostname != "third.invalid" || third.DefangedURL != "hxxps://third[.]invalid/final" {
		t.Fatalf("third redirect = %#v", third)
	}
	if len(analysis.NestedURLs) != 1 || analysis.NestedURLs[0].Hostname != "first.invalid" {
		t.Fatalf("first-hop compatibility view = %#v", analysis.NestedURLs)
	}
}

func TestURLPartsAnalysisMarksRedirectCyclesAndDepthLimit(t *testing.T) {
	cycleNodes := buildURLRedirectNodes(
		[]URLQueryParameter{{Key: "next", Value: "https://cycle.invalid/", HasValue: true}},
		2,
		map[string]bool{"https://cycle.invalid/": true},
	)
	if len(cycleNodes) != 1 || !cycleNodes[0].Cycle {
		t.Fatalf("cycle guard = %#v", cycleNodes)
	}

	value := "https://level5.invalid/"
	for level := 4; level >= 1; level-- {
		value = "https://level" + string(rune('0'+level)) + ".invalid/?next=" + urlQueryEscapeForTest(value)
	}
	depthDoc := mustParseURLPartsForTest(t, "https://root.invalid/?next="+urlQueryEscapeForTest(value))
	depthAnalysis, err := AnalyzeURLParts(depthDoc)
	if err != nil {
		t.Fatal(err)
	}
	node := depthAnalysis.RedirectChains[0]
	for len(node.Children) > 0 {
		node = node.Children[0]
	}
	if node.Depth != URLPartsRedirectMaxDepth || !node.Truncated {
		t.Fatalf("last redirect = %#v", node)
	}
}

func TestURLPartsReportsAreDefangedAndDeterministic(t *testing.T) {
	doc := mustParseURLPartsForTest(t, "https://root.invalid/?utm_source=mail&next=https%3A%2F%2Fchild.invalid%2Fpath")
	var first, second bytes.Buffer
	if err := EncodeURLPartsReportJSON(&first, doc); err != nil {
		t.Fatal(err)
	}
	if err := EncodeURLPartsReportJSON(&second, doc); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("JSON report is not deterministic")
	}
	var report URLPartsReport
	if err := json.Unmarshal(first.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.ReportVersion != 1 || report.URLPartsSchemaVersion != URLPartsSchemaVersion || report.DefangedURL != "hxxps://root[.]invalid/?utm_source=mail&next=hxxps%3A%2F%2Fchild%5B.%5Dinvalid%2Fpath" {
		t.Fatalf("report header = %#v", report)
	}
	if len(report.RedirectChains) != 1 || report.RedirectChains[0].DefangedURL != "hxxps://child[.]invalid/path" {
		t.Fatalf("report redirects = %#v", report.RedirectChains)
	}
	if strings.Contains(first.String(), "https://") || strings.Contains(first.String(), "http://") {
		t.Fatalf("JSON report leaked a live URL: %s", first.String())
	}
	markdown, err := URLPartsReportMarkdown(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# URL analysis report", "hxxps://root[.]invalid/", "utm\\_source", "hxxps://child\\[.\\]invalid/path", "not a safety verdict"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("Markdown report missing %q:\n%s", want, markdown)
		}
	}
}

func TestURLPartsPluginEmitsStandaloneReports(t *testing.T) {
	p := NewPluginURLParts()
	input := []byte("https://root.invalid/?next=https%3A%2F%2Fchild.invalid%2F")
	jsonReport := runFormat(t, p.Process, p.RegisterFlags, input, "-report", "json")
	if !bytes.Contains(jsonReport, []byte(`"report_version": 1`)) || bytes.Contains(jsonReport, []byte(`"url": "https://`)) {
		t.Fatalf("JSON report = %s", jsonReport)
	}
	markdown := runFormat(t, p.Process, p.RegisterFlags, input, "-report", "markdown")
	if !bytes.Contains(markdown, []byte("# URL analysis report")) || !bytes.Contains(markdown, []byte("hxxps://child\\[.\\]invalid/")) {
		t.Fatalf("Markdown report = %s", markdown)
	}
	if _, err := tryFormat(p.Process, p.RegisterFlags, input, "-report", "xml"); err == nil || !strings.Contains(err.Error(), "unsupported URL report format") {
		t.Fatalf("unsupported report error = %v", err)
	}
}

func mustParseURLPartsForTest(t *testing.T, rawURL string) *URLPartsDocument {
	t.Helper()
	doc, err := parseURLParts(strings.NewReader(rawURL))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func urlQueryEscapeForTest(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(value, "%", "%25"), ":", "%3A"), "/", "%2F")
}
