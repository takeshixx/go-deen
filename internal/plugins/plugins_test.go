package plugins

import (
	"sort"
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
	if !CmdAvailable("xor") {
		t.Error("Arithmetic plugin not found")
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

func TestInCategorySorted(t *testing.T) {
	names := InCategory("formatters")
	if len(names) == 0 {
		t.Fatal("expected formatter plugins")
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("formatter names are not sorted: %#v", names)
	}
}

func TestArithmeticCategory(t *testing.T) {
	names := InCategory("arithmetic")
	want := map[string]bool{"add": false, "not": false, "sub": false, "xor": false}
	for _, name := range names {
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("arithmetic category missing %q in %#v", name, names)
		}
	}
}

func TestCanDecode(t *testing.T) {
	if !CanDecode("base64") {
		t.Fatal("base64 should support decoding")
	}
	if CanDecode("sha256") {
		t.Fatal("sha256 should not support decoding")
	}
	if CanDecode("doesnotexist") {
		t.Fatal("unknown plugin should not support decoding")
	}
	if !CanDecode("xor") {
		t.Fatal("xor should support decoding")
	}
}
