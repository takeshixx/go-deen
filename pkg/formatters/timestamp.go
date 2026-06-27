package formatters

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

const defaultTimestampLayout = time.RFC3339Nano

// NewPluginTimestamp creates a reversible Unix timestamp formatter.
func NewPluginTimestamp() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "timestamp"
	p.Aliases = []string{".timestamp", "time", ".time", "ts", ".ts"}
	p.Category = "formatters"
	p.Description = "Convert Unix timestamps to formatted times and formatted times back to Unix timestamps."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("unit", "auto", "timestamp unit: auto, s, ms, us, ns")
		flags.String("layout", defaultTimestampLayout, "Go time layout for formatted output or parsing")
		flags.Bool("utc", true, "format parsed times in UTC")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := readTimestampInput(r)
		if err != nil {
			return err
		}
		n, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid Unix timestamp %q", input)
		}
		t, err := unixByUnit(n, helpers.StringFlag(flags, "unit"))
		if err != nil {
			return err
		}
		if helpers.IsBoolFlag(flags, "utc") {
			t = t.UTC()
		}
		_, err = io.WriteString(w, t.Format(helpers.StringFlag(flags, "layout")))
		return err
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		input, err := readTimestampInput(r)
		if err != nil {
			return err
		}
		t, err := parseTime(input, helpers.StringFlag(flags, "layout"))
		if err != nil {
			return err
		}
		if helpers.IsBoolFlag(flags, "utc") {
			t = t.UTC()
		}
		_, err = io.WriteString(w, unixStringByUnit(t, helpers.StringFlag(flags, "unit")))
		return err
	}
	return p
}

func readTimestampInput(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	input := strings.TrimSpace(string(data))
	if input == "" {
		return "", fmt.Errorf("empty timestamp input")
	}
	return input, nil
}

func unixByUnit(n int64, unit string) (time.Time, error) {
	switch normalizeTimestampUnit(unit, n) {
	case "s":
		return time.Unix(n, 0), nil
	case "ms":
		return time.Unix(0, n*int64(time.Millisecond)), nil
	case "us":
		return time.Unix(0, n*int64(time.Microsecond)), nil
	case "ns":
		return time.Unix(0, n), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported timestamp unit %q", unit)
	}
}

func unixStringByUnit(t time.Time, unit string) string {
	switch normalizeTimestampUnit(unit, 0) {
	case "ms":
		return strconv.FormatInt(t.UnixNano()/int64(time.Millisecond), 10)
	case "us":
		return strconv.FormatInt(t.UnixNano()/int64(time.Microsecond), 10)
	case "ns":
		return strconv.FormatInt(t.UnixNano(), 10)
	default:
		return strconv.FormatInt(t.Unix(), 10)
	}
}

func normalizeTimestampUnit(unit string, n int64) string {
	unit = strings.ToLower(strings.TrimSpace(unit))
	switch unit {
	case "", "auto":
		digits := len(strconv.FormatInt(absInt64(n), 10))
		switch {
		case digits >= 19:
			return "ns"
		case digits >= 16:
			return "us"
		case digits >= 13:
			return "ms"
		default:
			return "s"
		}
	case "sec", "secs", "second", "seconds":
		return "s"
	case "milli", "millis", "millisecond", "milliseconds":
		return "ms"
	case "micro", "micros", "microsecond", "microseconds":
		return "us"
	case "nano", "nanos", "nanosecond", "nanoseconds":
		return "ns"
	default:
		return unit
	}
}

func parseTime(input, layout string) (time.Time, error) {
	layouts := []string{
		layout,
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, candidate := range uniqueLayouts(layouts) {
		if t, err := time.Parse(candidate, input); err == nil {
			return t, nil
		}
	}
	for _, candidate := range uniqueLayouts(layouts) {
		if t, err := time.ParseInLocation(candidate, input, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse timestamp %q", input)
}

func uniqueLayouts(layouts []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, layout := range layouts {
		if layout == "" || seen[layout] {
			continue
		}
		seen[layout] = true
		out = append(out, layout)
	}
	return out
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
