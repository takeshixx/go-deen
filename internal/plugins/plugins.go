package plugins

import (
	"encoding/json"
	"fmt"
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
	hashs.NewPluginSHA512_224,
	hashs.NewPluginSHA512_256,
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
	hashs.NewPluginAdler32,
	hashs.NewPluginCRC32,
	hashs.NewPluginCRC32C,
	hashs.NewPluginCRC64ISO,
	hashs.NewPluginCRC64ECMA,
	hashs.NewPluginFNV32,
	hashs.NewPluginFNV32a,
	hashs.NewPluginFNV64,
	hashs.NewPluginFNV64a,
	hashs.NewPluginFNV128,
	hashs.NewPluginFNV128a,
	hashs.NewPluginHMAC,
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
// should be available accross all modules. The order
// determines how plugins are grouped in listings.
var PluginCategories = []string{"codecs", "compressions", "hashs", "formatters", "misc"}

// constructorByKey maps every plugin name and alias (with any leading "."
// stripped) to the constructor of the owning plugin. It is built once at
// package initialisation so command lookups are O(1) and do not reconstruct
// the whole plugin set on every call.
var constructorByKey = map[string]func() *types.DeenPlugin{}

// metadata holds a single instance of every plugin, built once for listings.
var metadata []*types.DeenPlugin

func init() {
	for _, constructor := range pluginConstructors {
		p := constructor()
		metadata = append(metadata, p)
		constructorByKey[lookupKey(p.Name)] = constructor
		for _, alias := range p.Aliases {
			constructorByKey[lookupKey(alias)] = constructor
		}
	}
}

// lookupKey normalises a command or alias to its canonical lookup form by
// dropping the leading "." that marks the unprocess (decode) direction.
func lookupKey(cmd string) string {
	return strings.TrimPrefix(cmd, ".")
}

// Resolve returns a fresh plugin instance for the given command along with the
// requested direction. A command prefixed with "." requests the unprocess
// (decode) direction. ok is false when no plugin matches.
func Resolve(cmd string) (plugin *types.DeenPlugin, unprocess bool, ok bool) {
	constructor, found := constructorByKey[lookupKey(cmd)]
	if !found {
		return nil, false, false
	}
	plugin = constructor()
	unprocess = strings.HasPrefix(cmd, ".")
	plugin.Command = lookupKey(cmd)
	return plugin, unprocess, true
}

type pluginDescription struct {
	Name    string
	Aliases []string
}

// PrintAvailable prints a list of available plugins
// and their aliases.
func PrintAvailable(outputJSON bool) {
	jsonObj := make(map[string][]pluginDescription)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', tabwriter.TabIndent)
	for _, category := range PluginCategories {
		if !outputJSON {
			fmt.Fprintf(w, "%s:\n", category)
		}
		for _, p := range metadata {
			if p.Category != category {
				continue
			}
			if outputJSON {
				jsonObj[category] = append(jsonObj[category], pluginDescription{
					Name:    p.Name,
					Aliases: p.Aliases,
				})
			} else if len(p.Aliases) > 0 {
				fmt.Fprintf(w, " \t%s\t%v\n", p.Name, p.Aliases)
			} else {
				fmt.Fprintf(w, " \t%s\n", p.Name)
			}
		}
		if !outputJSON {
			fmt.Fprintln(w, "")
		}
	}
	if outputJSON {
		encoded, err := json.Marshal(jsonObj)
		if err != nil {
			fmt.Fprintln(os.Stderr, "deen: failed to encode plugin list:", err)
			return
		}
		fmt.Println(string(encoded))
	} else {
		w.Flush()
	}
}

// CmdAvailable reports whether cmd is the name or alias of an available plugin.
func CmdAvailable(cmd string) bool {
	_, ok := constructorByKey[lookupKey(cmd)]
	return ok
}

// GetForCmd returns a fresh plugin instance for a given cmd, or nil if none
// matches. A "." prefix selects the unprocess direction.
func GetForCmd(cmd string) *types.DeenPlugin {
	plugin, _, ok := Resolve(cmd)
	if !ok {
		return nil
	}
	return plugin
}

// GetForCategory returns plugin names for a given category. When aliases is
// false only the plugin name (plus its "."-prefixed unprocess alias, if any)
// is returned; when true every alias is included.
func GetForCategory(category string, aliases bool) []string {
	var r []string
	for _, p := range metadata {
		if p.Category != category {
			continue
		}
		r = append(r, p.Name)
		if !aliases {
			if len(p.Aliases) > 0 && p.Aliases[0] == "."+p.Name {
				r = append(r, p.Aliases[0])
			}
		} else {
			r = append(r, p.Aliases...)
		}
	}
	return r
}
