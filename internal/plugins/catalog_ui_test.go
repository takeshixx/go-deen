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
