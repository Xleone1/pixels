package page

import (
	"fmt"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	outpage "github.com/niflaot/pixels/networking/outbound/catalog/page"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// imageSlotCount stores Nitro's fixed catalog image slot count.
	imageSlotCount = 3
	// textSlotCount stores Nitro's fixed catalog text slot count.
	textSlotCount = 4
)

// pageLocalization resolves every client layout slot with safe fallbacks.
func pageLocalization(translations i18n.Translator, page catalogmodel.Page) outpage.Localization {
	images := make([]string, imageSlotCount)
	texts := make([]string, textSlotCount)
	for index := range images {
		images[index] = optionalTranslation(translations, page.Name, fmt.Sprintf("image.%d", index))
	}
	for index := range texts {
		texts[index] = optionalTranslation(translations, page.Name, fmt.Sprintf("text.%d", index))
	}
	title := optionalTranslation(translations, page.Name, "")
	description := optionalTranslation(translations, page.Name, "description")
	applyLayoutFallbacks(page.Layout, title, description, texts)

	return outpage.Localization{Images: images, Texts: texts}
}

// applyLayoutFallbacks fills slots used by Nitro when no explicit copy exists.
func applyLayoutFallbacks(layout string, title string, description string, texts []string) {
	switch catalogmodel.NormalizeLayout(layout) {
	case catalogmodel.PetsInformationLayout:
		fillEmpty(&texts[1], title)
		fillEmpty(&texts[2], description)
	case catalogmodel.GroupFrontpageLayout:
		fillEmpty(&texts[0], description)
		fillEmpty(&texts[2], title)
	case catalogmodel.GroupForumLayout:
		fillEmpty(&texts[1], description)
	default:
		fillEmpty(&texts[0], description)
	}
}

// fillEmpty assigns fallback only when the explicit slot is empty.
func fillEmpty(target *string, fallback string) {
	if *target == "" {
		*target = fallback
	}
}

// optionalTranslation resolves one catalog page key without leaking missing keys.
func optionalTranslation(translations i18n.Translator, pageName string, suffix string) string {
	if translations == nil {
		return ""
	}
	key := "catalog.page." + pageName
	if suffix != "" {
		key += "." + suffix
	}
	resolved := translations.Default(i18n.Key(key))
	if resolved == key {
		return ""
	}

	return resolved
}
