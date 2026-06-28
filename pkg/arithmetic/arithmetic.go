package arithmetic

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/takeshixx/deen/pkg/types"
)

func byteValue(flags *flag.FlagSet, name string) (byte, error) {
	raw := strings.TrimSpace(flags.Lookup(name).Value.String())
	if len(raw) == 1 && (raw[0] < '0' || raw[0] > '9') {
		return raw[0], nil
	}
	n, err := strconv.ParseUint(raw, 0, 8)
	if err != nil {
		return 0, fmt.Errorf("%s must be a byte value such as 42, 0x2a, or '*'", name)
	}
	return byte(n), nil
}

func registerValueFlag(defaultValue, usage string) func(*flag.FlagSet) {
	return func(flags *flag.FlagSet) {
		flags.String("value", defaultValue, usage)
	}
}

func transformBytes(r io.Reader, w io.Writer, fn func(byte) byte) error {
	buf := make([]byte, 32*1024)
	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				buf[i] = fn(buf[i])
			}
			if _, err := w.Write(buf[:n]); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

// NewPluginXOR creates a byte-wise XOR plugin. XOR is symmetric, so processing
// and unprocessing apply the same transformation.
func NewPluginXOR() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "xor"
	p.Aliases = []string{".xor"}
	p.Category = "arithmetic"
	p.Description = "XOR every input byte with a byte value."
	p.RegisterFlags = registerValueFlag("0xff", "byte value to XOR with")
	transform := func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		value, err := byteValue(flags, "value")
		if err != nil {
			return err
		}
		return transformBytes(r, w, func(b byte) byte { return b ^ value })
	}
	p.Process = transform
	p.Unprocess = transform
	return p
}

// NewPluginAdd creates a byte-wise addition plugin. Decoding subtracts the
// configured value, wrapping at byte boundaries.
func NewPluginAdd() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "add"
	p.Aliases = []string{".add"}
	p.Category = "arithmetic"
	p.Description = "Add a byte value to every input byte, wrapping at 255."
	p.RegisterFlags = registerValueFlag("1", "byte value to add")
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		value, err := byteValue(flags, "value")
		if err != nil {
			return err
		}
		return transformBytes(r, w, func(b byte) byte { return b + value })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		value, err := byteValue(flags, "value")
		if err != nil {
			return err
		}
		return transformBytes(r, w, func(b byte) byte { return b - value })
	}
	return p
}

// NewPluginSub creates a byte-wise subtraction plugin. Decoding adds the
// configured value, wrapping at byte boundaries.
func NewPluginSub() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "sub"
	p.Aliases = []string{".sub"}
	p.Category = "arithmetic"
	p.Description = "Subtract a byte value from every input byte, wrapping at 0."
	p.RegisterFlags = registerValueFlag("1", "byte value to subtract")
	p.Process = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		value, err := byteValue(flags, "value")
		if err != nil {
			return err
		}
		return transformBytes(r, w, func(b byte) byte { return b - value })
	}
	p.Unprocess = func(r io.Reader, w io.Writer, flags *flag.FlagSet) error {
		value, err := byteValue(flags, "value")
		if err != nil {
			return err
		}
		return transformBytes(r, w, func(b byte) byte { return b + value })
	}
	return p
}

// NewPluginNot creates a byte-wise bitwise NOT plugin. NOT is symmetric.
func NewPluginNot() *types.DeenPlugin {
	p := types.NewPlugin()
	p.Name = "not"
	p.Aliases = []string{".not"}
	p.Category = "arithmetic"
	p.Description = "Invert every input byte with bitwise NOT."
	transform := func(r io.Reader, w io.Writer, _ *flag.FlagSet) error {
		return transformBytes(r, w, func(b byte) byte { return ^b })
	}
	p.Process = transform
	p.Unprocess = transform
	return p
}
