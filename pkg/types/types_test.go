package types

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()
	if reflect.TypeOf(p) != reflect.TypeOf(&DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPlugin: %s", reflect.TypeOf(p))
	}
}

func TestTrimReader(t *testing.T) {
	testData := "   white spaces at the beginning and end	"
	testReader := strings.NewReader(testData)
	var trimReader TrimReader
	trimReader.Rd = testReader
	var destBuf bytes.Buffer
	n, err := io.Copy(&destBuf, trimReader)
	if err != nil {
		t.Error(err)
	}
	if int(n) != len(strings.TrimSpace(testData)) {
		t.Error(errors.New("Invalid number of bytes copied"))
	}
	if strings.TrimSpace(testData) != destBuf.String() {
		t.Error(errors.New("Invalid data returned"))
	}
}

func TestDeenTask(t *testing.T) {
	var w bytes.Buffer
	task := NewDeenTask(&w)
	err := task.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestDeenTaskError(t *testing.T) {
	var w bytes.Buffer
	testErr := errors.New("a test error")
	task := NewDeenTask(&w)
	go func() {
		select {
		case err := <-task.ErrChan:
			if err != testErr {
				t.Errorf("Invalid error returned: %v != %v", testErr, err)
			}
		case <-task.DoneChan:
		}
	}()
	task.Error(testErr)
	err := task.Close()
	if err != nil {
		t.Error(err)
	}
}
