package types

import (
	"flag"
	"io"
)

type PluginConstructor func() DeenPlugin
type FuncStub func([]byte) []byte
type StreamFuncStub func(io.Reader) ([]byte, error)
type AddCliFuncStub func(*DeenPlugin, []string) *flag.FlagSet
type FuncWithCliFlagsStub func(*flag.FlagSet, []byte) ([]byte, error)
type StreamFuncWithCliFlagsStub func(*flag.FlagSet, io.Reader) ([]byte, error)

type DeenPlugin struct {
	Name                            string
	Aliases                         []string
	Type                            string
	Unprocess                       bool
	ProcessFunc                     FuncStub
	UnprocessFunc                   FuncStub
	ProcessStreamFunc               StreamFuncStub
	UnprocessStreamFunc             StreamFuncStub
	ProcessWithCliFlagsFunc         FuncWithCliFlagsStub
	ProcessStreamWithCliFlagsFunc   StreamFuncWithCliFlagsStub
	UnprocessStreamWithCliFlagsFunc StreamFuncWithCliFlagsStub
	AddCliOptionsFunc               AddCliFuncStub
	CliHelp                         string
}
