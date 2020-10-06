package codecs

import (
	"bytes"
	"encoding/base64"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/takeshixx/deen/pkg/helpers"
	"github.com/takeshixx/deen/pkg/types"
)

var certPEM = `-----BEGIN CERTIFICATE-----
MIIFVzCCBD+gAwIBAgISBKHwjurPCnhEk9DJ5AabAPL9MA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDEwMDYwOTE3MzBaFw0y
MTAxMDQwOTE3MzBaMBMxETAPBgNVBAMTCGRlZW4ub29vMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEAtbG8b6Hesb0wgDE9cLM8xmuqJmuMtgIKMjVGxTW3
+ipQwUWDzuw9QGqB6wrneIqwBa8j+4qM9WLW/NvjA0XtbaeCkSw1foe8BDWE5mK6
6qCQN9awlAzPv7tFqu/UqL8DcgIVLCZJfTA8Yx7b9VLlejjbaPa4gApBCZo3jZRx
fSMfarezWN99YP9LTIzkiYZoG36I87pSxX8ZT4cvifh8T2ArYILb96I5lmiFF99P
I62hTJXHa2q2CUC0iCVDued//BzwY6fuZgRZADuPwKAWJ8oLnAXu1J4THc+lU5em
l9ilg29DFqnKZRQbjwuj+6WzgzeI0ESPUqwkSpS4byttMwIDAQABo4ICbDCCAmgw
DgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBQEtZf8cquSaVGCezrY98Jc8h94NzAfBgNV
HSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcBAQRjMGEwLgYI
KwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlwdC5vcmcwLwYI
KwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlwdC5vcmcvMCEG
A1UdEQQaMBiCCGRlZW4ub29vggx3d3cuZGVlbi5vb28wTAYDVR0gBEUwQzAIBgZn
gQwBAgEwNwYLKwYBBAGC3xMBAQEwKDAmBggrBgEFBQcCARYaaHR0cDovL2Nwcy5s
ZXRzZW5jcnlwdC5vcmcwggEFBgorBgEEAdZ5AgQCBIH2BIHzAPEAdgBc3EOS/uar
RUSxXprUVuYQN/vV+kfcoXOUsl7m9scOygAAAXT9aoafAAAEAwBHMEUCIFLmiYMS
neur1J4hbNBeDy7QZXTnfcTD1zuM0wKVE3PkAiEAgeK9+zztyXbr5e8rTmmjouw4
uawbnoEq+LjZY4v0gC0AdwB9PvL4j/+IVWgkwsDKnlKJeSvFDngJfy5ql2iZfiLw
1wAAAXT9aobFAAAEAwBIMEYCIQCJiNWYbwxfmqLsPdDEQWnwT5agh5uJVAps7Pzw
4NX67gIhAJIdx8f9skOEjlZIRUG4Fr/Bb0XxCW4tn0TqdOjEm9FvMA0GCSqGSIb3
DQEBCwUAA4IBAQBapvJMD0PGcEuK6nEJPCu086mNGH0s0SxbTL383WcleD9OQP8X
zb1XDYjpiDnipOw3la630TyRXW5K3gTWHiGsymoyJZ7poOXlpB1SdGUldzhB4Nef
N7wzMNh/ergEtic5MjgHl7IsIzA+mWAv2CiAOw4eumVpE1zWZ//vyofszw8tHBML
7Xi3bywE4AxpT9bcFWCPjX6S+/xdmB9CzgzZQCATxc8+MG8ZlpgnYjSgn98XGxPJ
Edzk7AAvHgTDUK6zHNuAs7N19fLoi5aUr1wacamhemGd3aCTc/uS3ZA/oIMwXdF9
4adHiE9HYQqkUg+QyZwajblJK+GsGgyJnQb5Cg==
-----END CERTIFICATE-----
`
var certDERB64 = "MIIFVzCCBD+gAwIBAgISBKHwjurPCnhEk9DJ5AabAPL9MA0GCSqGSIb3DQEBCwUAMEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQDExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0yMDEwMDYwOTE3MzBaFw0yMTAxMDQwOTE3MzBaMBMxETAPBgNVBAMTCGRlZW4ub29vMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtbG8b6Hesb0wgDE9cLM8xmuqJmuMtgIKMjVGxTW3+ipQwUWDzuw9QGqB6wrneIqwBa8j+4qM9WLW/NvjA0XtbaeCkSw1foe8BDWE5mK66qCQN9awlAzPv7tFqu/UqL8DcgIVLCZJfTA8Yx7b9VLlejjbaPa4gApBCZo3jZRxfSMfarezWN99YP9LTIzkiYZoG36I87pSxX8ZT4cvifh8T2ArYILb96I5lmiFF99PI62hTJXHa2q2CUC0iCVDued//BzwY6fuZgRZADuPwKAWJ8oLnAXu1J4THc+lU5eml9ilg29DFqnKZRQbjwuj+6WzgzeI0ESPUqwkSpS4byttMwIDAQABo4ICbDCCAmgwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBQEtZf8cquSaVGCezrY98Jc8h94NzAfBgNVHSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcBAQRjMGEwLgYIKwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlwdC5vcmcwLwYIKwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlwdC5vcmcvMCEGA1UdEQQaMBiCCGRlZW4ub29vggx3d3cuZGVlbi5vb28wTAYDVR0gBEUwQzAIBgZngQwBAgEwNwYLKwYBBAGC3xMBAQEwKDAmBggrBgEFBQcCARYaaHR0cDovL2Nwcy5sZXRzZW5jcnlwdC5vcmcwggEFBgorBgEEAdZ5AgQCBIH2BIHzAPEAdgBc3EOS/uarRUSxXprUVuYQN/vV+kfcoXOUsl7m9scOygAAAXT9aoafAAAEAwBHMEUCIFLmiYMSneur1J4hbNBeDy7QZXTnfcTD1zuM0wKVE3PkAiEAgeK9+zztyXbr5e8rTmmjouw4uawbnoEq+LjZY4v0gC0AdwB9PvL4j/+IVWgkwsDKnlKJeSvFDngJfy5ql2iZfiLw1wAAAXT9aobFAAAEAwBIMEYCIQCJiNWYbwxfmqLsPdDEQWnwT5agh5uJVAps7Pzw4NX67gIhAJIdx8f9skOEjlZIRUG4Fr/Bb0XxCW4tn0TqdOjEm9FvMA0GCSqGSIb3DQEBCwUAA4IBAQBapvJMD0PGcEuK6nEJPCu086mNGH0s0SxbTL383WcleD9OQP8Xzb1XDYjpiDnipOw3la630TyRXW5K3gTWHiGsymoyJZ7poOXlpB1SdGUldzhB4NefN7wzMNh/ergEtic5MjgHl7IsIzA+mWAv2CiAOw4eumVpE1zWZ//vyofszw8tHBML7Xi3bywE4AxpT9bcFWCPjX6S+/xdmB9CzgzZQCATxc8+MG8ZlpgnYjSgn98XGxPJEdzk7AAvHgTDUK6zHNuAs7N19fLoi5aUr1wacamhemGd3aCTc/uS3ZA/oIMwXdF94adHiE9HYQqkUg+QyZwajblJK+GsGgyJnQb5Cg=="

func TestNewPluginPEM(t *testing.T) {
	p := NewPluginPEM()
	if reflect.TypeOf(p) != reflect.TypeOf(&types.DeenPlugin{}) {
		t.Errorf("Invalid return type for NewPluginPEM: %s", reflect.TypeOf(p))
	}
}

func TestPluginPEMProcessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	decodedDER, err := base64.StdEncoding.DecodeString(certDERB64)
	if err != nil {
		t.Error(err)
	}
	task.Reader = bytes.NewReader(decodedDER)
	plugin := NewPluginPEM()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{"-cert"})
	plugin.ProcessDeenTaskWithFlags(flags, task)
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), []byte(certPEM)); c != 0 {
			t.Errorf("TestPluginPEMProcessDeenTask data wrong: %s != %s", destWriter.Bytes(), []byte(certPEM))
		}
	}
}

func TestPluginPEMUnprocessDeenTask(t *testing.T) {
	destWriter := new(bytes.Buffer)
	task := types.NewDeenTask(destWriter)
	task.Reader = strings.NewReader(certPEM)
	plugin := NewPluginPEM()
	flags := helpers.DefaultFlagSet()
	flags = plugin.AddDefaultCliFunc(plugin, flags, []string{})
	plugin.UnprocessDeenTaskWithFlags(flags, task)
	decodedDER, err := base64.StdEncoding.DecodeString(certDERB64)
	if err != nil {
		t.Error(err)
	}
	select {
	case err := <-task.ErrChan:
		t.Error(err)
	case <-task.DoneChan:
		if c := bytes.Compare(destWriter.Bytes(), decodedDER); c != 0 {
			t.Errorf("TestPluginPEMUnprocessDeenTask data wrong: %s != %s", destWriter.Bytes(), decodedDER)
		}
	}
}

func TestPluginPEMUsage(t *testing.T) {
	p := NewPluginPEM()
	flags := helpers.DefaultFlagSet()
	flags = p.AddDefaultCliFunc(p, flags, []string{})
	_, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stderr = w
	flags.Usage()
}
