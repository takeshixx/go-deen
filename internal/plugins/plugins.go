package plugins

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/takeshixx/deen/pkg/codecs"
	"github.com/takeshixx/deen/pkg/compressions"
	"github.com/takeshixx/deen/pkg/formatters"
	"github.com/takeshixx/deen/pkg/hashs"
	"github.com/takeshixx/deen/pkg/types"
)

var pluginConstructors = []types.PluginConstructor{
	codecs.NewPluginBase32,
	codecs.NewPluginBase32Hex,
	codecs.NewPluginBase64,
	codecs.NewPluginBase85,
	codecs.NewPluginHex,
	codecs.NewPluginURL,
	codecs.NewPluginHTML,
	codecs.NewPluginJwt,
	hashs.NewPluginSHA1,
	hashs.NewPluginSHA224,
	hashs.NewPluginSHA256,
	hashs.NewPluginSHA384,
	hashs.NewPluginSHA512,
	hashs.NewPluginMD4,
	hashs.NewPluginMD5,
	hashs.NewPluginRIPEMD160,
	hashs.NewPluginBLAKE2s,
	hashs.NewPluginBLAKE2s128,
	hashs.NewPluginBLAKE2b,
	hashs.NewPluginBLAKE2b384,
	hashs.NewPluginBLAKE2b256,
	hashs.NewPluginBLAKE3,
	compressions.NewPluginFlate,
	compressions.NewPluginLZMA,
	compressions.NewPluginLZMA2,
	compressions.NewPluginLzw,
	compressions.NewPluginGzip,
	compressions.NewPluginZlib,
	compressions.NewPluginBzip2,
	formatters.NewPluginJSONFormatter,
}

var PluginCategories = []string{"codec", "compression", "hash", "formatter"}

// PrintAvailable prints a list of available plugins
// and their aliases.
func PrintAvailable() {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	for _, constructor := range pluginConstructors {
		p := constructor()
		if len(p.Aliases) > 0 {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", p.Name, p.Aliases))
		} else {
			fmt.Fprintln(w, p.Name)
		}
	}
	w.Flush()
}

// CmdAvailable checks if the given cmd is the name or
// or a alias of an available plugin.
func CmdAvailable(cmd string) bool {
	for _, constructor := range pluginConstructors {
		p := constructor()
		if cmd == p.Name {
			return true
		} else {
			for _, alias := range p.Aliases {
				if alias == cmd {
					return true
				}
			}
		}
	}
	return false
}

// GetForCmd returns the plugin object for a given cmd
func GetForCmd(cmd string) (plugin *types.DeenPlugin) {
	for _, constructor := range pluginConstructors {
		p := constructor()
		if p.Name == cmd {
			plugin = &p
			break
		} else {
			for _, alias := range p.Aliases {
				if alias == cmd {
					plugin = &p
					break
				}
			}
		}
	}
	if plugin != nil && strings.HasPrefix(cmd, ".") {
		plugin.Unprocess = true
	}
	// TODO: the function should return an error object
	return
}

// GetForCategory returns plugin names for a given category
func GetForCategory(category string, aliases bool) []string {
	var r []string
	for _, constructor := range pluginConstructors {
		p := constructor()
		if p.Type == category {
			r = append(r, p.Name)
			if aliases == false {
				if len(p.Aliases) > 0 {
					if p.Aliases[0] == "."+p.Name {
						r = append(r, p.Aliases[0])
					}
				}
			} else if aliases && len(p.Aliases) > 0 {
				for _, alias := range p.Aliases {
					r = append(r, alias)
				}
			}
		}
	}
	return r
}
