package formatters

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestProtobufFormatter(t *testing.T) {
	p := NewPluginProtobuf()
	var out bytes.Buffer
	input := []byte{
		0x08, 0x96, 0x01, // field 1, varint 150
		0x12, 0x04, 't', 'e', 's', 't', // field 2, string "test"
	}
	if err := p.Process(bytes.NewReader(input), &out, flag.NewFlagSet("protobuf", flag.ContinueOnError)); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"1: varint 150", `2: string "test"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestProtobufFormatterNestedMessage(t *testing.T) {
	p := NewPluginProtobuf()
	var out bytes.Buffer
	input := []byte{
		0x0a, 0x02, // field 1, length 2
		0x08, 0x01, // nested field 1, varint 1
	}
	if err := p.Process(bytes.NewReader(input), &out, flag.NewFlagSet("protobuf", flag.ContinueOnError)); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"1: message 2 bytes {", "  1: varint 1", "}"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestProtobufFormatterRejectsInvalidInput(t *testing.T) {
	p := NewPluginProtobuf()
	var out bytes.Buffer
	if err := p.Process(bytes.NewReader([]byte{0x00}), &out, flag.NewFlagSet("protobuf", flag.ContinueOnError)); err == nil {
		t.Fatal("expected invalid field number error")
	}
}
