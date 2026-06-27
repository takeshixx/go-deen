package misc

import (
	"flag"
	"fmt"
	"io"
	"regexp"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginRegex creates a regex extract and replace plugin.
func NewPluginRegex() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "regex"
	p.Aliases = []string{"re"}
	p.Category = "misc"
	p.Description = "Extract regex matches or replace text using Go regular expressions."
	p.RegisterFlags = func(flags *flag.FlagSet) {
		flags.String("re", "", "regular expression")
		flags.String("replace", "", "replacement string; when set, performs replacement")
		flags.Int("group", 0, "capture group to extract")
		flags.Bool("all", true, "extract all matches")
	}
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		pattern := helpers.StringFlag(flags, "re")
		if pattern == "" {
			return fmt.Errorf("missing -re")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		input, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		repl := helpers.StringFlag(flags, "replace")
		if repl != "" {
			_, err = w.Write(re.ReplaceAll(input, []byte(repl)))
			return err
		}
		group := helpers.IntFlag(flags, "group", 0)
		matches := re.FindAllSubmatch(input, -1)
		if !helpers.IsBoolFlag(flags, "all") && len(matches) > 1 {
			matches = matches[:1]
		}
		for i, match := range matches {
			if group < 0 || group >= len(match) {
				return fmt.Errorf("group %d unavailable", group)
			}
			if i > 0 {
				fmt.Fprintln(w)
			}
			w.Write(match[group])
		}
		return nil
	}
	return p
}
