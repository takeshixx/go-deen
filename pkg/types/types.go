package types

import (
	"flag"
	"io"
	"log"
)

type PluginConstructor func() DeenPlugin
type AddCliFuncStub func(*DeenPlugin, []string) *flag.FlagSet

type StreamFuncStub func(io.Reader) ([]byte, error)
type StreamFuncWithCliFlagsStub func(*flag.FlagSet, io.Reader) ([]byte, error)

type DeenTaskFuncStub func(*DeenTask)
type DeenTaskWithFlagsStub func(*flag.FlagSet, *DeenTask)

type PipeChanFuncStub func(io.Reader, *io.PipeWriter, chan bool, chan error)
type PipeChanFuncWithFlagsStub func(*flag.FlagSet, io.Reader, *io.PipeWriter, chan bool, chan error)

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

	ProcessDeenTaskFunc        DeenTaskFuncStub
	UnprocessDeenTaskFunc      DeenTaskFuncStub
	ProcessDeenTaskWithFlags   DeenTaskWithFlagsStub
	UnprocessDeenTaskWithFlags DeenTaskWithFlagsStub

	ProcessPipeChanFunc        PipeChanFuncStub
	UnprocessPipeChanFunc      PipeChanFuncStub
	ProcessPipeChanWithFlags   PipeChanFuncWithFlagsStub
	UnprocessPipeChanWithFlags PipeChanFuncWithFlagsStub
	AddCliOptionsFunc          AddCliFuncStub
	CliHelp                    string
}

type DeenTask struct {
	Reader     io.Reader
	PipeReader *io.PipeReader
	PipeWriter *io.PipeWriter
	DoneChan   chan bool
	ErrChan    chan error
}

func (dt *DeenTask) Close() error {
	err := dt.PipeWriter.Close()
	dt.DoneChan <- true
	return err
}

func (dt *DeenTask) Error(err error) {
	dt.ErrChan <- err
}

func NewDeenTask(writer io.Writer) *DeenTask {
	dt := &DeenTask{}
	pr, pw := io.Pipe()
	dt.PipeReader = pr
	dt.PipeWriter = pw
	dt.DoneChan = make(chan bool)
	dt.ErrChan = make(chan error)
	go func() {
		defer close(dt.DoneChan)
		_, err := io.Copy(writer, dt.PipeReader)
		if err != nil {
			log.Fatalf("Reading from plugin pipe failed: %v\n", err)
		}
	}()
	return dt
}
