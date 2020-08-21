package helpers

import (
	"flag"
	"strconv"
)

// IsBoolFlag returns the boolean value of a FlagSet flag.
func IsBoolFlag(flags *flag.FlagSet, name string) (cur bool) {
	curFlag := flags.Lookup(name)
	cur, err := strconv.ParseBool(curFlag.Value.String())
	if err != nil {
		cur = false
	}
	return
}

// DefaultFlagSet creates a default FlagSet with
// predefined options that should bbe available
// globally.
func DefaultFlagSet() *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.String("file", "", "read from file")
	return flags
}

// RemoveBeforeSubcommand removes all elements from slice inArr
// before the string cmd (including cmd) and returns a new slice
// without those elements.
func RemoveBeforeSubcommand(inArr []string, cmd string) []string {
	var outArr []string
	for i, e := range inArr {
		if e == cmd {
			outArr = inArr[i+1:]
		}
	}
	return outArr
}
