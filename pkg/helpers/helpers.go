package helpers

import (
	"flag"
	"strconv"
)

// IsBoolFlag returns the boolean value of a FlagSet flag. It returns false if
// the flag is not present or does not hold a boolean.
func IsBoolFlag(flags *flag.FlagSet, name string) bool {
	if flags == nil {
		return false
	}
	f := flags.Lookup(name)
	if f == nil {
		return false
	}
	cur, err := strconv.ParseBool(f.Value.String())
	if err != nil {
		return false
	}
	return cur
}

// IntFlag returns the integer value of a FlagSet flag, or def if the flag is
// absent or not a valid integer.
func IntFlag(flags *flag.FlagSet, name string, def int) int {
	if flags == nil {
		return def
	}
	f := flags.Lookup(name)
	if f == nil {
		return def
	}
	v, err := strconv.Atoi(f.Value.String())
	if err != nil {
		return def
	}
	return v
}

// StringFlag returns the string value of a FlagSet flag, or "" if absent.
func StringFlag(flags *flag.FlagSet, name string) string {
	if flags == nil {
		return ""
	}
	f := flags.Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

// DefaultFlagSet creates a default FlagSet with
// predefined options that should bbe available
// globally.
func DefaultFlagSet() *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Bool("n", false, "do not output the trailing newline")
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
