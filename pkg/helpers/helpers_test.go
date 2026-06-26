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

func TestIntFlag(t *testing.T) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.Int("level", 5, "level")
	flags.Parse([]string{"-level", "9"})
	if got := IntFlag(flags, "level", 1); got != 9 {
		t.Errorf("IntFlag = %d, want 9", got)
	}
	if got := IntFlag(flags, "missing", 7); got != 7 {
		t.Errorf("IntFlag default = %d, want 7", got)
	}
}

func TestStringFlag(t *testing.T) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.String("key", "", "key")
	flags.Parse([]string{"-key", "value"})
	if got := StringFlag(flags, "key"); got != "value" {
		t.Errorf("StringFlag = %q, want value", got)
	}
}

func TestRemoveBeforeSubcommand(t *testing.T) {
	inArr := []string{"path", "-a", "mycmd", "-file", "test"}
	testData := RemoveBeforeSubcommand(inArr, "mycmd")
	if reflect.DeepEqual(testData, inArr[2:]) {
		t.Errorf("Invalid return data: %v\n", testData)
	}
}
