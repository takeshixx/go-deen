package formatters

import (
	"strings"
	"testing"
)

func TestCSVFormatsTable(t *testing.T) {
	p := NewPluginCSV()
	out := runFormat(t, p.Process, p.RegisterFlags, []byte("name,ok\nadeen,true\n"))
	got := string(out)
	for _, want := range []string{"name", "ok", "adeen", "true"} {
		if !strings.Contains(got, want) {
			t.Fatalf("CSV table missing %q:\n%s", want, got)
		}
	}
}

func TestCSVToTSV(t *testing.T) {
	p := NewPluginCSV()
	out := runFormat(t, p.Process, p.RegisterFlags, []byte("name,ok\nadeen,true\n"), "-out", "tsv")
	if got, want := string(out), "name\tok\nadeen\ttrue\n"; got != want {
		t.Fatalf("CSV to TSV = %q, want %q", got, want)
	}
}

func TestTSVToCSV(t *testing.T) {
	p := NewPluginCSV()
	out := runFormat(t, p.Unprocess, p.RegisterFlags, []byte("name\tok\nadeen\ttrue\n"), "-in", "tsv")
	if got, want := string(out), "name,ok\nadeen,true\n"; got != want {
		t.Fatalf("TSV to CSV = %q, want %q", got, want)
	}
}

func TestCSVRejectsBadDelimiter(t *testing.T) {
	p := NewPluginCSV()
	if _, err := tryFormat(p.Process, p.RegisterFlags, []byte("a,b\n"), "-in", "bad"); err == nil {
		t.Fatal("expected bad delimiter to fail")
	}
}
