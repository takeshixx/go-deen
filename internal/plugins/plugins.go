package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/takeshixx/deen/pkg/codecs"
	"github.com/takeshixx/deen/pkg/compressions"
	"github.com/takeshixx/deen/pkg/formatters"
	"github.com/takeshixx/deen/pkg/hashs"
	"github.com/takeshixx/deen/pkg/misc"
	"github.com/takeshixx/deen/pkg/types"
)

var pluginConstructors = []func() *types.DeenPlugin{
	codecs.NewPluginBase32,
	codecs.NewPluginBase64,
	codecs.NewPluginBase85,
	codecs.NewPluginHex,
	codecs.NewPluginURL,
	codecs.NewPluginHTML,
	codecs.NewPluginUnicode,
	codecs.NewPluginStrconv,
	codecs.NewPluginPEM,
	hashs.NewPluginSHA1,
	hashs.NewPluginSHA224,
	hashs.NewPluginSHA256,
	hashs.NewPluginSHA384,
	hashs.NewPluginSHA512,
	hashs.NewPluginSHA3224,
	hashs.NewPluginSHA3256,
	hashs.NewPluginSHA3384,
	hashs.NewPluginSHA3512,
	hashs.NewPluginMD4,
	hashs.NewPluginMD5,
	hashs.NewPluginRIPEMD160,
	hashs.NewPluginBLAKE2s,
	hashs.NewPluginBLAKE2b,
	hashs.NewPluginBLAKE2x,
	hashs.NewPluginBLAKE3,
	hashs.NewPluginBcrypt,
	hashs.NewPluginScrypt,
	compressions.NewPluginFlate,
	compressions.NewPluginLZMA,
	compressions.NewPluginLZMA2,
	compressions.NewPluginLzw,
	compressions.NewPluginGzip,
	compressions.NewPluginZlib,
	compressions.NewPluginBzip2,
	compressions.NewPluginBrotli,
	formatters.NewPluginJSONFormatter,
	formatters.NewPluginJwt,
	formatters.NewPluginJQFormatter,
	misc.NewPluginCertCloner,
	misc.NewPluginCertPrinter,
}

// PluginCategories is a list of plugin categories that
// should be available accross all modules.
var PluginCategories = []string{"codecs", "compressions", "hashs", "formatters", "utils", "misc"}

type pluginDescription struct {
	Name    string
	Aliases []string
}

// PrintAvailable prints a list of available plugins
// and their aliases.
func PrintAvailable(outputJSON bool) {
	var pluginList []*types.DeenPlugin
	for _, constructor := range pluginConstructors {
		p := constructor()
		pluginList = append(pluginList, p)
	}
	var jsonObj map[string][]pluginDescription
	jsonObj = make(map[string][]pluginDescription)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', tabwriter.TabIndent)
	for _, category := range PluginCategories {
		if !outputJSON {
			fmt.Fprintf(w, "%s:\n", category)
		}
		for _, p := range pluginList {
			if p.Category != category {
				continue
			}
			if len(p.Aliases) > 0 {
				if outputJSON {
					jsonObj[category] = append(jsonObj[category], pluginDescription{
						Name:    p.Name,
						Aliases: []string{},
					})
				} else {
					fmt.Fprintf(w, " \t%s\t%s\n", p.Name, p.Aliases)
				}
			} else {
				if outputJSON {
					jsonObj[category] = append(jsonObj[category], pluginDescription{
						Name:    p.Name,
						Aliases: []string{},
					})
				} else {
					fmt.Fprintf(w, " \t%s\n", p.Name)
				}
			}
		}
		if !outputJSON {
			fmt.Fprintln(w, "")
		}
	}
	if outputJSON {
		encoded, err := json.Marshal(jsonObj)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(encoded))
	} else {
		w.Flush()
	}
}

// CmdAvailable checks if the given cmd is the name or
// or a alias of an available plugin.
func CmdAvailable(cmd string) bool {
	for _, constructor := range pluginConstructors {
		p := constructor()
		if cmd == p.Name {
			return true
		}
		for _, alias := range p.Aliases {
			if alias == cmd {
				return true
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
			plugin = p
			break
		} else {
			for _, alias := range p.Aliases {
				if alias == cmd {
					plugin = p
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
		if p.Category == category {
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
