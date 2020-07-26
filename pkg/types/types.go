package types

import (
	"flag"
	"io"
)

type PluginConstructor func() DeenPlugin
type AddCliFuncStub func(*DeenPlugin, []string) *flag.FlagSet

type StreamFuncStub func(io.Reader) ([]byte, error)
type StreamFuncWithCliFlagsStub func(*flag.FlagSet, io.Reader) ([]byte, error)

type PipeFuncStub func(io.Reader, *io.PipeWriter) error
type PipeFuncWithFlagsStub func(*flag.FlagSet, io.Reader, *io.PipeWriter) error

type DeenPlugin struct {
	Name                            string
	Aliases                         []string
	Type                            string
	Unprocess                       bool
	ProcessStreamFunc               StreamFuncStub
	UnprocessStreamFunc             StreamFuncStub
	ProcessStreamWithCliFlagsFunc   StreamFuncWithCliFlagsStub
	UnprocessStreamWithCliFlagsFunc StreamFuncWithCliFlagsStub
	ProcessPipeFunc                 PipeFuncStub
	UnprocessPipeFunc               PipeFuncStub
	ProcessPipeWithFlags            PipeFuncWithFlagsStub
	UnprocessPipeWithFlags          PipeFuncWithFlagsStub
	AddCliOptionsFunc               AddCliFuncStub
	CliHelp                         string
}
