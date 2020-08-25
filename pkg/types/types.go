package types

import (
	"bytes"
	"flag"
	"io"
	"log"
)

type PluginConstructor func() DeenPlugin
type AddDefaultCliFunc func(*DeenPlugin, *flag.FlagSet, []string) *flag.FlagSet

type StreamFuncStub func(io.Reader) ([]byte, error)
type StreamFuncWithCliFlagsStub func(*flag.FlagSet, io.Reader) ([]byte, error)

type DeenTaskFuncStub func(*DeenTask)
type DeenTaskWithFlagsStub func(*flag.FlagSet, *DeenTask)

type DeenPlugin struct {
	Name              string
	Aliases           []string
	Type              string
	Unprocess         bool
	AddDefaultCliFunc AddDefaultCliFunc
	CliHelp           string

	ProcessDeenTaskFunc        DeenTaskFuncStub
	UnprocessDeenTaskFunc      DeenTaskFuncStub
	ProcessDeenTaskWithFlags   DeenTaskWithFlagsStub
	UnprocessDeenTaskWithFlags DeenTaskWithFlagsStub

	// TODO: port and remove old stream funcs
	ProcessStreamFunc               StreamFuncStub
	UnprocessStreamFunc             StreamFuncStub
	ProcessStreamWithCliFlagsFunc   StreamFuncWithCliFlagsStub
	UnprocessStreamWithCliFlagsFunc StreamFuncWithCliFlagsStub
}

// DeenTask describes a (un)processing
// task for a plugin. It includes pointers
// to all relevant inputs and outputs.
type DeenTask struct {
	Reader     io.Reader
	PipeReader *io.PipeReader
	PipeWriter *io.PipeWriter
	DoneChan   chan bool
	ErrChan    chan error
}

// Close is responsible for closing all
// relevant chans and pipes of the
// corresponding task.
func (dt *DeenTask) Close() error {
	err := dt.PipeWriter.Close()
	// TODO: should we also send to DoneChan?
	//dt.DoneChan <- true
	return err
}

func (dt *DeenTask) Error(err error) {
	dt.ErrChan <- err
}

// NewDeenTask creates a new task that can be used
// for processing/unprocessing funcs of plugins.
// It also makes it easier to setup test cases.
func NewDeenTask(writer io.Writer) *DeenTask {
	dt := &DeenTask{}
	// pipeWriter is provided to plugins to write data to.
	// After a plugin has finished, processed data will be
	// piped from pipeReader to os.Stdout
	pr, pw := io.Pipe()
	dt.PipeReader = pr
	dt.PipeWriter = pw
	dt.DoneChan = make(chan bool)
	dt.ErrChan = make(chan error)
	// We have to ensure that the pipeReader reads data
	// so that the pipeWriter does not block. After Copy
	// is done or pipeWriter is closed it will also
	// close the doneChan
	go func() {
		defer close(dt.DoneChan)
		_, err := io.Copy(writer, dt.PipeReader)
		if err != nil {
			log.Fatalf("Reading from plugin pipe failed: %v\n", err)
		}
	}()
	return dt
}

type TrimReader struct {
	Rd io.Reader
}

func (tr TrimReader) Read(buf []byte) (int, error) {
	n, err := tr.Rd.Read(buf)
	t := bytes.TrimSpace(buf[:n])
	n = copy(buf, t)
	return n, err
}
