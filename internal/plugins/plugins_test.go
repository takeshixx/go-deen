package plugins

import (
	"testing"
)

func TestPrintAvailable(t *testing.T) {
	PrintAvailable(false)
	PrintAvailable(true)
}

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

func TestGetForCategory(t *testing.T) {
	validCategory := GetForCategory("codecs", true)
	if validCategory == nil || len(validCategory) < 1 {
		t.Error("No plugins found for valid category")
	}

	validCategory = GetForCategory("codecs", false)
	if validCategory == nil || len(validCategory) < 1 {
		t.Error("No plugins found for valid category")
	}

	invalidCategory := GetForCategory("nocategory", true)
	if len(invalidCategory) > 0 {
		t.Error("Invalid category returned plugin list")
	}

	invalidCategory = GetForCategory("nocategory", false)
	if len(invalidCategory) > 0 {
		t.Error("Invalid category returned plugin list")
	}
}
