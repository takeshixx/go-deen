package types

import (
	"bytes"
	"flag"
	"io"
	"log"
)

// DeenPlugin is the base struct for
// each plugin instance.
type DeenPlugin struct {
	Name                            string
	Aliases                         []string
	Category                        string
	Unprocess                       bool
	CliHelp                         string
	Command                         string // The command with which the plugin was called
	AddDefaultCliFunc               func(*DeenPlugin, *flag.FlagSet, []string) *flag.FlagSet
	ProcessDeenTaskFunc             func(*DeenTask)
	UnprocessDeenTaskFunc           func(*DeenTask)
	ProcessDeenTaskWithFlags        func(*flag.FlagSet, *DeenTask)
	UnprocessDeenTaskWithFlags      func(*flag.FlagSet, *DeenTask)
	ProcessStreamFunc               func(io.Reader) ([]byte, error)
	UnprocessStreamFunc             func(io.Reader) ([]byte, error)
	ProcessStreamWithCliFlagsFunc   func(*flag.FlagSet, io.Reader) ([]byte, error)
	UnprocessStreamWithCliFlagsFunc func(*flag.FlagSet, io.Reader) ([]byte, error)
}

// NewPlugin creates a plugin skeleton.
func NewPlugin() *DeenPlugin {
	p := &DeenPlugin{}
	return p
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
	Command    string // The command with which the plugin was called
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

// TrimReader creates a reader that
// trims the input while reading.
type TrimReader struct {
	Rd io.Reader
}

func (tr TrimReader) Read(buf []byte) (int, error) {
	n, err := tr.Rd.Read(buf)
	t := bytes.TrimSpace(buf[:n])
	n = copy(buf, t)
	return n, err
}
