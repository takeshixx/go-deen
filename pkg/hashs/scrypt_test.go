package hashs

import (
	"fmt"
	"testing"
)

var scryptTestData = []byte("verysecurepassword")
var scryptTestSaltHex = "7465737473616c74"

func TestPluginScrypt(t *testing.T) {
	assertHash(t, NewPluginScrypt(), scryptTestData, "0IlaUn16JVhuciJDk6ow51TFVcgG7T14dYlJsNsWFRQ=")
	assertHash(t, NewPluginScrypt(), scryptTestData, "m39r9JZGQxVP76+lu4zGiFCbspLci15u9OTI9zlC1lU=", "-salt", scryptTestSaltHex)
	assertHash(t, NewPluginScrypt(), scryptTestData,
		"CLf5VMJVqyLztPE4cK1fpoRbOwQDUSgYm4VWfxgdMvpH6dBbaJ2rD0+hRhC6vcbEaL0/XHQSJTFYifshfoIRh+B2RRhZKqeTpXqP+4jxhiuMVa1lgInQMlAflOQfSCaq",
		"-cost", fmt.Sprintf("%d", 1<<12), "-len", "96", "-salt", scryptTestSaltHex)
}
