package types

import (
	"flag"
	"io"
)

// TransformFunc is the unified entry point for a plugin operation. It reads
// input from r, writes the transformed result to w and returns any error.
// Implementations must return (not continue) on the first error.
type TransformFunc func(r io.Reader, w io.Writer, flags *flag.FlagSet) error

// DeenPlugin describes a single encode/decode/hash/format operation.
type DeenPlugin struct {
	Name        string
	Aliases     []string
	Category    string
	Description string

	// RegisterFlags registers plugin-specific CLI flags. It may be nil.
	RegisterFlags func(*flag.FlagSet)
	// Process performs the forward operation (encode/compress/hash/format).
	Process TransformFunc
	// Unprocess performs the reverse operation (decode/decompress). A nil
	// value means the plugin is one-way (e.g. hashes).
	Unprocess TransformFunc

	// Command is the command (alias) with which the plugin was invoked, with
	// any leading "." stripped. Plugins that expose several aliases (e.g. the
	// unicode plugin) use it to pick a variant.
	Command string
}

// NewPlugin creates an empty plugin skeleton.
func NewPlugin() *DeenPlugin {
	return &DeenPlugin{}
}
