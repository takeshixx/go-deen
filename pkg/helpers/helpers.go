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
