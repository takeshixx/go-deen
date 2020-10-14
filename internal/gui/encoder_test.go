// +build gui

package gui

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewDeenEncoderWidget(t *testing.T) {
	g, err := NewDeenGUI()
	if err != nil {
		t.Error(err)
	}
	e, err := NewDeenEncoderWidget(g)
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(e) != reflect.TypeOf(&DeenEncoder{}) {
		t.Errorf("Invalid type returned: %v\n", e)
	}
}

func TestEncoderSetContent(t *testing.T) {
	g, err := NewDeenGUI()
	if err != nil {
		t.Error(err)
	}
	e, err := NewDeenEncoderWidget(g)
	if err != nil {
		t.Error(err)
	}

	testData := []byte("this is a test!!")
	e.SetContent(testData)
	if bytes.Compare(e.Content, testData) != 0 {
		t.Errorf("Encoder content is different: %v != %v\n", e.Content, testData)
	}
	if bytes.Compare(e.GetContent(), testData) != 0 {
		t.Errorf("Encoder text is different: %v != %v\n", e.Content, testData)
	}
}

func TestEncoderTypeText(t *testing.T) {
	g, err := NewDeenGUI()
	if err != nil {
		t.Error(err)
	}
	e, err := NewDeenEncoderWidget(g)
	if err != nil {
		t.Error(err)
	}

	testData := []byte("this is a test!!")
	e.InputField.SetText(string(testData))
	if bytes.Compare(e.GetContent(), testData) != 0 {
		t.Errorf("Encoder text is different: %v != %v\n", e.Content, testData)
	}

	e.InputField.SetText("")
	if bytes.Compare(e.GetContent(), []byte("")) != 0 {
		t.Errorf("Returned invalid data: %v\n", e.GetContent())
	}
}

func TestNewDeenInputField(t *testing.T) {
	p := &DeenEncoder{}
	f := NewDeenInputField(p)
	if reflect.TypeOf(f) != reflect.TypeOf(&DeenInputField{}) {
		t.Errorf("Invalid type returned: %v\n", f)
	}
}
