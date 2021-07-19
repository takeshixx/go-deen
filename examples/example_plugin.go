package examples

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

// NewPluginExample creates a new example plugin instance that
// should be used for all plugins that implement readers/writers.
// This typically applies to plugins that do not have a fixed-size
// output, e.g. codecs or compressions.
func NewPluginExample() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "example"
	p.Aliases = []string{".example", "ex", ".ex"}
	p.Category = "codecs"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		// Processing of DeenTasks is done in goroutines
		go func() {
			// Example of processing DeenTasks by applying base64 encoding
			processor := base64.NewEncoder(base64.StdEncoding, task.PipeWriter)
			_, err := io.Copy(processor, task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copying into processor failed")
			}
			err = processor.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing processor failed")
			}
			// Make sure to close the PipeWriter at the end.
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter failed")
			}
		}()
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			// Example of unprocessing by base64 decoding input
			processor := base64.NewDecoder(base64.StdEncoding, task.Reader)
			_, err := io.Copy(task.PipeWriter, processor)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in processor failed")
			}
			// Make sure to close the PipeWriter at the end.
			err = task.PipeWriter.Close()
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Closing PipeWriter failed")
			}
		}()
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		// Read string option from flags
		strFlag := flags.Lookup("teststr")
		strVal := strFlag.Value.String()
		if strVal != "" {
			// String option has been set
		}

		// Read bool option from flags with IsBoolFlag()
		boolVal := helpers.IsBoolFlag(flags, "testbool")
		if boolVal == true {
			// Bool option has been set
		}

		// In case there are no additional options for this plugin,
		// just redirect to the ProcessDeenTask func:
		// p.ProcessDeenTaskFunc(task)

		// Otherwise, implement processing with additional options.

	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		// See ProcessDeenTaskWithFlags()
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			// Add a description for the plugin.
			fmt.Fprintf(os.Stderr, "Plugin description...\n\n")
			// Additional arguments will be listed automatically
			// in the help page, they should not be mentioned in
			// the above description.
			flags.PrintDefaults()
		}
		// Adding additional flags:
		//flags.String("example", "", "example string option")

		// Different options for processing and unprocessing
		// can be added by checking:
		//if self.Unprocess {
		//
		//}

		flags.Parse(args)
		return flags
	}
	return
}

// NewPluginStreamExample creates a stream-based plugin that can
// be used for plugins that do not implement readers/writers.
// This often applies to plugins that have fixed-size outputs
// like hashs, that return byte arrays instead of writing to
// writers directly. These types of plugins might also not
// have a unprocessing functions because they implement a
// one-way process.
func NewPluginStreamExample() (p types.DeenPlugin) {
	p.Name = "streamexample"
	p.Aliases = []string{}
	p.Category = "hashs"
	p.Unprocess = false
	p.ProcessStreamFunc = func(reader io.Reader) ([]byte, error) {
		// Example processing with SHA1
		var err error
		hasher := sha1.New()
		if _, err := io.Copy(hasher, reader); err != nil {
			return nil, err
		}
		hashSum := hasher.Sum(nil)
		return hashSum, err
	}
	p.ProcessStreamWithCliFlagsFunc = func(flags *flag.FlagSet, reader io.Reader) ([]byte, error) {
		// Read string option from flags
		strFlag := flags.Lookup("teststr")
		strVal := strFlag.Value.String()
		if strVal != "" {
			// String option has been set
		}

		// Read bool option from flags with IsBoolFlag()
		boolVal := helpers.IsBoolFlag(flags, "testbool")
		if boolVal == true {
			// Bool option has been set
		}

		// In case there are no additional options for this plugin,
		// just redirect to the ProcessDeenTask func:
		// p.ProcessDeenTaskFunc(task)

		// Otherwise, implement processing with additional options.
		return nil, errors.New("Default error")
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			// Add a description for the plugin.
			fmt.Fprintf(os.Stderr, "Plugin description...\n\n")
			// Additional arguments will be listed automatically
			// in the help page, they should not be mentioned in
			// the above description.
			flags.PrintDefaults()
		}
		// Adding additional flags:
		//flags.String("example", "", "example string option")

		// Different options for processing and unprocessing
		// can be added by checking:
		//if self.Unprocess {
		//
		//}

		flags.Parse(args)
		return flags
	}
	return
}
