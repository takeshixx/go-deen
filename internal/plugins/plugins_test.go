package plugins

import (
	"testing"
)

func TestCmdAvailable(t *testing.T) {
	if !CmdAvailable("html") {
		t.Error("HTML plugin not found")
	}
	if CmdAvailable("doesnotexist") {
		t.Error("Found a plugin that does not exist")
	}
	if !CmdAvailable("b85") {
		t.Error("Did not find plugin with aliases")
	}
}

/* func TestGetForCmd(t *testing.T) {
	pa := GetForCmd("base64")
	pb := codecs.NewPluginBase64()
	if !reflect.DeepEqual(pa, pb) {
		t.Errorf("GetForCmd did not return a proper plugin: %v != %v", reflect.TypeOf(pa), reflect.TypeOf(pb))
	}

	pa = GetForCmd("b85")
	pb = codecs.NewPluginBase64()
	if reflect.TypeOf(pa) != reflect.TypeOf(&pb) {
		t.Errorf("GetForCmd did not return a proper plugin: %v != %v", reflect.TypeOf(pa), reflect.TypeOf(pb))
	}

	pa = GetForCmd(".brotli")
	pb = codecs.NewPluginBase64()
	if reflect.TypeOf(pa) != reflect.TypeOf(&pb) {
		t.Errorf("GetForCmd did not return a proper plugin: %v != %v", reflect.TypeOf(pa), reflect.TypeOf(pb))
	}
}
*/
