package pipeline

import (
	"flag"

	"github.com/takeshixx/deen/internal/plugins"
)

// Option describes a single configurable plugin flag for UI rendering.
type Option struct {
	Name    string
	Usage   string
	Default string
	IsBool  bool
}

// PluginOptions returns the configurable options (flags) of a plugin, or nil if
// it has none. Bool flags are reported with IsBool so the UI can render a
// checkbox instead of a text entry.
func PluginOptions(name string) []Option {
	p, _, ok := plugins.Resolve(name)
	if !ok || p.RegisterFlags == nil {
		return nil
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	p.RegisterFlags(fs)

	var opts []Option
	fs.VisitAll(func(f *flag.Flag) {
		_, isBool := f.Value.(interface{ IsBoolFlag() bool })
		opts = append(opts, Option{
			Name:    f.Name,
			Usage:   f.Usage,
			Default: f.DefValue,
			IsBool:  isBool,
		})
	})
	return opts
}
