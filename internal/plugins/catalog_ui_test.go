package plugins

import "testing"

func TestUICatalogHasUserFacingCopy(t *testing.T) {
	catalog := UICatalog()
	if len(catalog) == 0 {
		t.Fatal("UICatalog returned no plugins")
	}
	for _, info := range catalog {
		if info.Description == "" {
			t.Errorf("%s has no description", info.Name)
		}
		if info.UseFor == "" {
			t.Errorf("%s has no use-case copy", info.Name)
		}
		for _, ref := range info.References {
			if ref.Label == "" || ref.URL == "" {
				t.Errorf("%s has an incomplete reference: %#v", info.Name, ref)
			}
		}
	}
}

func TestSearchUICatalog(t *testing.T) {
	if got := SearchUICatalog(""); len(got) != len(UICatalog()) {
		t.Fatalf("empty search returned %d plugins, want %d", len(got), len(UICatalog()))
	}

	var foundBase64 bool
	for _, info := range SearchUICatalog("b64") {
		if info.Name == "base64" {
			foundBase64 = true
		}
	}
	if !foundBase64 {
		t.Fatal("alias search did not find base64")
	}

	var foundJWT bool
	for _, info := range SearchUICatalog("authentication tokens") {
		if info.Name == "jwt" {
			foundJWT = true
		}
	}
	if !foundJWT {
		t.Fatal("use-case search did not find jwt")
	}
}
