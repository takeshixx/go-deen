package formatters

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// URLPartsReportVersion is the standalone analysis report schema version.
const URLPartsReportVersion = 1

// URLPartsReport is a deterministic, defanged summary suitable for sharing.
type URLPartsReport struct {
	ReportVersion         int                         `json:"report_version"`
	URLPartsSchemaVersion int                         `json:"url_parts_schema_version"`
	DefangedURL           string                      `json:"defanged_url"`
	Indicators            []URLPartsIndicator         `json:"indicators"`
	TrackingParameters    []URLPartsTrackingParameter `json:"tracking_parameters"`
	RedirectChains        []URLPartsReportRedirect    `json:"redirect_chains"`
}

// URLPartsReportRedirect omits the live URL and retains its defanged form.
type URLPartsReportRedirect struct {
	Depth          int                      `json:"depth"`
	ParameterIndex int                      `json:"parameter_index"`
	Key            string                   `json:"key"`
	DefangedURL    string                   `json:"defanged_url"`
	Scheme         string                   `json:"scheme"`
	Hostname       string                   `json:"hostname"`
	Cycle          bool                     `json:"cycle"`
	Truncated      bool                     `json:"truncated"`
	Children       []URLPartsReportRedirect `json:"children"`
}

// BuildURLPartsReport derives a standalone report without trusting a stored
// analysis object and without performing network access.
func BuildURLPartsReport(doc *URLPartsDocument) (*URLPartsReport, error) {
	analysis, err := AnalyzeURLParts(doc)
	if err != nil {
		return nil, err
	}
	return &URLPartsReport{
		ReportVersion:         URLPartsReportVersion,
		URLPartsSchemaVersion: URLPartsSchemaVersion,
		DefangedURL:           analysis.DefangedURL,
		Indicators:            append([]URLPartsIndicator{}, analysis.Indicators...),
		TrackingParameters:    append([]URLPartsTrackingParameter{}, analysis.TrackingParameters...),
		RedirectChains:        reportRedirects(analysis.RedirectChains),
	}, nil
}

// EncodeURLPartsReportJSON writes the standalone report as readable JSON.
func EncodeURLPartsReportJSON(w io.Writer, doc *URLPartsDocument) error {
	report, err := BuildURLPartsReport(doc)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	return enc.Encode(report)
}

// URLPartsReportMarkdown returns a deterministic, defanged Markdown report.
func URLPartsReportMarkdown(doc *URLPartsDocument) (string, error) {
	report, err := BuildURLPartsReport(doc)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString("# URL analysis report\n\n")
	out.WriteString("Local observations only; no destinations were contacted. This is not a safety verdict.\n\n")
	out.WriteString("## Defanged URL\n\n    ")
	out.WriteString(strings.ReplaceAll(report.DefangedURL, "\n", " "))
	out.WriteString("\n\n## Indicators\n\n")
	if len(report.Indicators) == 0 {
		out.WriteString("- No common indicators detected.\n")
	} else {
		for _, indicator := range report.Indicators {
			fmt.Fprintf(&out, "- **%s** `%s`: %s\n", strings.ToUpper(indicator.Severity), markdownEscape(indicator.Code), markdownEscape(indicator.Message))
		}
	}
	out.WriteString("\n## Common tracking parameters\n\n")
	if len(report.TrackingParameters) == 0 {
		out.WriteString("- None detected.\n")
	} else {
		for _, tracking := range report.TrackingParameters {
			fmt.Fprintf(&out, "- Query #%d: `%s`\n", tracking.ParameterIndex, markdownEscape(tracking.Key))
		}
	}
	out.WriteString("\n## Local redirect chains\n\n")
	if len(report.RedirectChains) == 0 {
		out.WriteString("- None detected.\n")
	} else {
		writeMarkdownRedirects(&out, report.RedirectChains, 0)
	}
	return out.String(), nil
}

func reportRedirects(nodes []URLPartsRedirectNode) []URLPartsReportRedirect {
	report := make([]URLPartsReportRedirect, len(nodes))
	for i, node := range nodes {
		report[i] = URLPartsReportRedirect{
			Depth:          node.Depth,
			ParameterIndex: node.ParameterIndex,
			Key:            node.Key,
			DefangedURL:    node.DefangedURL,
			Scheme:         node.Scheme,
			Hostname:       node.Hostname,
			Cycle:          node.Cycle,
			Truncated:      node.Truncated,
			Children:       reportRedirects(node.Children),
		}
	}
	return report
}

func writeMarkdownRedirects(out *strings.Builder, nodes []URLPartsReportRedirect, indent int) {
	for _, node := range nodes {
		status := ""
		if node.Cycle {
			status = " — cycle detected"
		} else if node.Truncated {
			status = " — deeper values omitted"
		}
		fmt.Fprintf(out, "%s- Query #%d (`%s`) → %s%s\n", strings.Repeat("  ", indent), node.ParameterIndex, markdownEscape(node.Key), markdownEscape(node.DefangedURL), status)
		writeMarkdownRedirects(out, node.Children, indent+1)
	}
}

func markdownEscape(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	replacer := strings.NewReplacer(
		"\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "#", "\\#", "<", "\\<", ">", "\\>",
	)
	return replacer.Replace(value)
}
