//go:build webembed

// Package web provides the deen WebAssembly interface assets. When built with
// the "webembed" tag they are embedded into the binary so `deen serve` is
// self-contained; otherwise FS returns nil and a directory must be served.
package web

import (
	"embed"
	"io/fs"
)

//go:embed assets
var assets embed.FS

// Embedded reports whether the web assets are compiled into this binary.
const Embedded = true

// FS returns the embedded web assets, or nil if they are unavailable.
func FS() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil
	}
	return sub
}
