package codecs

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/takeshixx/deen/pkg/types"
)

func parseHeaders(flags *flag.FlagSet) (ret map[string]string, err error) {
	headersPtr := flags.Lookup("headers")
	headers := headersPtr.Value.String()
	if headers == "" {
		ret = make(map[string]string)
		return
	}
	err = json.Unmarshal([]byte(headers), &ret)
	return
}

func encodePEM(t *types.DeenTask, dataType string, headers map[string]string) {
	go func() {
		defer t.Close()
		data, err := ioutil.ReadAll(t.Reader)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Reading in PEM failed")
		}
		block := &pem.Block{
			Type:    dataType,
			Headers: headers,
			Bytes:   data,
		}
		err = pem.Encode(t.PipeWriter, block)
		if err != nil {
			t.ErrChan <- errors.Wrap(err, "Copying into encoder in PEM failed")
		}
	}()
}

// NewPluginPEM creates a new PluginPEM object
func NewPluginPEM() (p *types.DeenPlugin) {
	p = types.NewPlugin()
	p.Name = "pem"
	p.Aliases = []string{".pem"}
	p.Category = "codecs"
	p.Unprocess = false
	p.ProcessDeenTaskFunc = func(task *types.DeenTask) {
		encodePEM(task, "MESSAGE", make(map[string]string))
	}
	p.ProcessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		headers, err := parseHeaders(flags)
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse headers")
			}()
			return
		}
		certFlag := flags.Lookup("cert")
		cert, err := strconv.ParseBool(certFlag.Value.String())
		if err != nil {
			go func() {
				defer task.Close()
				task.ErrChan <- errors.Wrap(err, "Failed to parse cert flag")
			}()
			return
		}
		dataType := "MESSAGE"
		if cert {
			dataType = "CERTIFICATE"
		}
		encodePEM(task, dataType, headers)
	}
	p.UnprocessDeenTaskFunc = func(task *types.DeenTask) {
		go func() {
			defer task.Close()
			data, err := ioutil.ReadAll(task.Reader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Reading in PEM failed")
			}
			block, _ := pem.Decode(data)
			blockReader := bytes.NewReader(block.Bytes)
			_, err = io.Copy(task.PipeWriter, blockReader)
			if err != nil {
				task.ErrChan <- errors.Wrap(err, "Copy in PEM failed")
			}
		}()
	}
	p.UnprocessDeenTaskWithFlags = func(flags *flag.FlagSet, task *types.DeenTask) {
		p.UnprocessDeenTaskFunc(task)
	}
	p.AddDefaultCliFunc = func(self *types.DeenPlugin, flags *flag.FlagSet, args []string) *flag.FlagSet {
		flags.Init(p.Name, flag.ExitOnError)
		flags.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", p.Name)
			fmt.Fprintf(os.Stderr, "Privacy Enhanced Mail (PEM) data encoding/decoding. (RFC 1421)\n\n")
			flags.PrintDefaults()
		}
		if !self.Unprocess {
			flags.String("type", "MESSAGE", "data type")
			flags.String("headers", "", "message headers in JSON format")
			flags.Bool("cert", false, "create a PEM encoded certificate")
		}
		flags.Parse(args)
		return flags
	}
	return
}
