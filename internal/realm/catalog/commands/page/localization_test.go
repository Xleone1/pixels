package page

import (
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestPageLocalizationMapsExplicitSlots verifies private catalogs can fill every Nitro slot.
func TestPageLocalizationMapsExplicitSlots(t *testing.T) {
	t.Parallel()
	entries := map[i18n.Key]string{
		"catalog.page.custom.image.0": "headline.png",
		"catalog.page.custom.image.1": "teaser.png",
		"catalog.page.custom.image.2": "special.png",
		"catalog.page.custom.text.0":  "intro",
		"catalog.page.custom.text.1":  "body",
		"catalog.page.custom.text.2":  "details",
		"catalog.page.custom.text.3":  "footer",
	}
	translations := i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{"es": entries})
	localization := pageLocalization(translations, catalogmodel.Page{Name: "custom", Layout: catalogmodel.PetsInformationLayout})
	for index, expected := range []string{"headline.png", "teaser.png", "special.png"} {
		if localization.Images[index] != expected {
			t.Fatalf("image %d=%q, want %q", index, localization.Images[index], expected)
		}
	}
	for index, expected := range []string{"intro", "body", "details", "footer"} {
		if localization.Texts[index] != expected {
			t.Fatalf("text %d=%q, want %q", index, localization.Texts[index], expected)
		}
	}
}

// TestPageLocalizationUsesLayoutFallbacks verifies every rich layout receives visible copy.
func TestPageLocalizationUsesLayoutFallbacks(t *testing.T) {
	t.Parallel()
	translations := i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{
		"es": {
			"catalog.page.custom":             "Título",
			"catalog.page.custom.description": "Descripción",
		},
	})
	cases := []struct {
		layout string
		index  int
		value  string
	}{
		{catalogmodel.DefaultLayout, 0, "Descripción"},
		{catalogmodel.InformationLayout, 0, "Descripción"},
		{catalogmodel.PetsInformationLayout, 1, "Título"},
		{catalogmodel.PetsInformationLayout, 2, "Descripción"},
		{catalogmodel.GroupFrontpageLayout, 0, "Descripción"},
		{catalogmodel.GroupFrontpageLayout, 2, "Título"},
		{catalogmodel.GroupForumLayout, 1, "Descripción"},
	}
	for _, test := range cases {
		localization := pageLocalization(translations, catalogmodel.Page{Name: "custom", Layout: test.layout})
		if localization.Texts[test.index] != test.value {
			t.Fatalf("layout %q text %d=%q, want %q", test.layout, test.index, localization.Texts[test.index], test.value)
		}
	}
}

// TestPageLocalizationDoesNotExposeMissingKeys verifies optional copy remains empty.
func TestPageLocalizationDoesNotExposeMissingKeys(t *testing.T) {
	t.Parallel()
	localization := pageLocalization(i18n.NewCatalog(i18n.Config{}, nil), catalogmodel.Page{Name: "private", Layout: catalogmodel.PetsInformationLayout})
	for index, value := range localization.Texts {
		if value != "" {
			t.Fatalf("text %d leaked %q", index, value)
		}
	}
}
