package formatters

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

func delimiter(name string) (rune, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "csv", ",":
		return ',', nil
	case "tsv", "tab", "\\t":
		return '\t', nil
	case "semicolon", ";":
		return ';', nil
	default:
		if len([]rune(name)) == 1 {
			return []rune(name)[0], nil
		}
		return 0, fmt.Errorf("unsupported delimiter %q", name)
	}
}

func readDelimited(r io.Reader, delim rune) ([][]string, error) {
	cr := csv.NewReader(r)
	cr.Comma = delim
	cr.FieldsPerRecord = -1
	return cr.ReadAll()
}

func writeDelimited(w io.Writer, rows [][]string, delim rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
	if err := cw.WriteAll(rows); err != nil {
		return err
	}
	return cw.Error()
}

func writeTable(w io.Writer, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// NewPluginCSV creates a CSV/TSV table formatter and delimiter converter.
func NewPluginCSV() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "csv"
	p.Aliases = []string{"tsv"}
	p.Category = "formatters"
	p.Description = "Format CSV/TSV as a readable table and convert between delimiters."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("in", "csv", "input delimiter: csv, tsv, semicolon or a single character")
		flags.String("out", "table", "output format: table, csv, tsv, semicolon or a single character")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		inDelim, err := delimiter(helpers.StringFlag(flags, "in"))
		if err != nil {
			return err
		}
		rows, err := readDelimited(r, inDelim)
		if err != nil {
			return err
		}
		out := strings.ToLower(strings.TrimSpace(helpers.StringFlag(flags, "out")))
		if out == "" || out == "table" {
			return writeTable(w, rows)
		}
		outDelim, err := delimiter(out)
		if err != nil {
			return err
		}
		return writeDelimited(w, rows, outDelim)
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		in := helpers.StringFlag(flags, "in")
		if in == "" {
			in = "tsv"
		}
		inDelim, err := delimiter(in)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		rows, err := readDelimited(bytes.NewReader(data), inDelim)
		if err != nil {
			return err
		}
		return writeDelimited(w, rows, ',')
	}
	return p
}
