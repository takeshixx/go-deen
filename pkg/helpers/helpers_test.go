package helpers

import (
	"flag"
	"reflect"
	"testing"
)

func TestIsBoolFlag(t *testing.T) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.Bool("test", false, "test value")
	flags.Parse([]string{"-test"})
	testVal := IsBoolFlag(flags, "test")
	if testVal != true {
		t.Error("Failed to parse true bool from FlagSet")
	}

	flags = flag.NewFlagSet("", flag.ContinueOnError)
	flags.Bool("test", false, "test value")
	flags.Parse([]string{})
	testVal = IsBoolFlag(flags, "test")
	if testVal != false {
		t.Error("Failed to parse false bool from FlagSet")
	}
}

func TestDefaultFlagSet(t *testing.T) {
	flags := DefaultFlagSet()
	if flags.Lookup("file") == nil {
		t.Error("Could not find file option in FlagSet")
	}
}

func TestRemoveBeforeSubcommand(t *testing.T) {
	inArr := []string{"path", "-a", "mycmd", "-file", "test"}
	testData := RemoveBeforeSubcommand(inArr, "mycmd")
	if reflect.DeepEqual(testData, inArr[2:]) {
		t.Errorf("Invalid return data: %v\n", testData)
	}
}
