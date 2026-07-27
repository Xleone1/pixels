package catalog

import "testing"

// TestLoadConfigReadsFreeItems verifies the debug pricing startup switch.
func TestLoadConfigReadsFreeItems(t *testing.T) {
	t.Setenv("PIXELS_CATALOG_FREE_ITEMS", "true")

	config, err := LoadConfig()
	if err != nil || !config.FreeItems {
		t.Fatalf("config=%#v error=%v", config, err)
	}
}

// TestLoadConfigDefaultsToPaidCatalog verifies production-safe pricing.
func TestLoadConfigDefaultsToPaidCatalog(t *testing.T) {
	t.Setenv("PIXELS_CATALOG_FREE_ITEMS", "")

	config, err := LoadConfig()
	if err != nil || config.FreeItems {
		t.Fatalf("config=%#v error=%v", config, err)
	}
}
