package types

import (
	"flag"
	"io"
)

type PluginConstructor func() DeenPlugin
type AddCliFuncStub func(*DeenPlugin, []string) *flag.FlagSet

type StreamFuncStub func(io.Reader) ([]byte, error)
type StreamFuncWithCliFlagsStub func(*flag.FlagSet, io.Reader) ([]byte, error)

type PipeFuncStub func(io.Reader, *io.Writer) error
type PipeFuncWithCLIFlagsStub func(*flag.FlagSet, io.Reader, *io.Writer) error

type DeenPlugin struct {
	Name                            string
	Aliases                         []string
	Type                            string
	Unprocess                       bool
	ProcessStreamFunc               StreamFuncStub
	UnprocessStreamFunc             StreamFuncStub
	ProcessStreamWithCliFlagsFunc   StreamFuncWithCliFlagsStub
	UnprocessStreamWithCliFlagsFunc StreamFuncWithCliFlagsStub
	AddCliOptionsFunc               AddCliFuncStub
	CliHelp                         string
}
