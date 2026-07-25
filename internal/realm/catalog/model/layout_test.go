package model

import "testing"

// TestNormalizeLayoutMapsEveryLegacyHubblyLayout verifies stable Nitro compatibility.
func TestNormalizeLayoutMapsEveryLegacyHubblyLayout(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"club_buy":            VIPBuyLayout,
		"club_gift":           ClubGiftsLayout,
		"collectibles":        DefaultLayout,
		"frontpage":           InformationLayout,
		"frontpage_featured":  InformationLayout,
		"guild_furni":         GroupCustomFurnitureLayout,
		"guilds":              GroupFrontpageLayout,
		"info_duckets":        InformationLayout,
		"info_pets":           InformationLayout,
		"info_rentables":      InformationLayout,
		"loyalty_vip_buy":     VIPBuyLayout,
		"petcustomization":    DefaultLayout,
		"plasto":              ColorGroupingLayout,
		"productpage1":        SingleBundleLayout,
		"recycler":            InformationLayout,
		"recycler_info":       InformationLayout,
		"recycler_prizes":     DefaultLayout,
		"spaces":              SpacesLayout,
		SoundMachineLayout:    SoundMachineLayout,
		GroupFrontpageLayout:  GroupFrontpageLayout,
		RecentPurchasesLayout: RecentPurchasesLayout,
	}
	for source, expected := range cases {
		if actual := NormalizeLayout(source); actual != expected {
			t.Fatalf("layout %q normalized to %q, want %q", source, actual, expected)
		}
	}
}
