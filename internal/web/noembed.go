//go:build !webembed

// Package web provides the deen WebAssembly interface assets. Without the
// "webembed" build tag no assets are embedded and FS returns nil, so callers
// must serve a directory instead.
package web

import "io/fs"

// Embedded reports whether the web assets are compiled into this binary.
const Embedded = false

// FS returns the embedded web assets, or nil if they are unavailable.
func FS() fs.FS { return nil }
