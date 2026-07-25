package model

const (
	// BadgeDisplayLayout identifies the badge-display furniture editor.
	BadgeDisplayLayout = "badge_display"
	// BotsLayout identifies catalog offers that grant inventory bots.
	BotsLayout = "bots"
	// ClubGiftsLayout identifies the monthly club-gift selector.
	ClubGiftsLayout = "club_gifts"
	// ColorGroupingLayout identifies furniture with selectable color variants.
	ColorGroupingLayout = "default_3x3_color_grouping"
	// GroupCustomFurnitureLayout identifies group-customizable furniture.
	GroupCustomFurnitureLayout = "guild_custom_furni"
	// GroupForumLayout identifies group forum purchases.
	GroupForumLayout = "guild_forum"
	// GroupFrontpageLayout identifies the group creation landing page.
	GroupFrontpageLayout = "guild_frontpage"
	// InformationLayout identifies scrollable catalog information pages.
	InformationLayout = "info_loyalty"
	// MarketplaceLayout identifies public marketplace searches.
	MarketplaceLayout = "marketplace"
	// MarketplaceOwnItemsLayout identifies a player's marketplace offers.
	MarketplaceOwnItemsLayout = "marketplace_own_items"
	// PetsLayout identifies typed pet purchase pages.
	PetsLayout = "pets"
	// PetsInformationLayout identifies rich pet and catalog information pages.
	PetsInformationLayout = "pets3"
	// RoomAdsLayout identifies room-promotion purchases.
	RoomAdsLayout = "roomads"
	// RoomBundleLayout identifies complete room bundles.
	RoomBundleLayout = "room_bundle"
	// SingleBundleLayout identifies one highlighted bundle.
	SingleBundleLayout = "single_bundle"
	// SoundMachineLayout identifies the intentionally unsupported jukebox catalog.
	SoundMachineLayout = "soundmachine"
	// SpacesLayout identifies wall, floor, and landscape decorators.
	SpacesLayout = "spaces_new"
	// TrophiesLayout identifies customizable trophy offers.
	TrophiesLayout = "trophies"
	// VIPBuyLayout identifies club subscription offers.
	VIPBuyLayout = "vip_buy"
)

// NormalizeLayout maps legacy Hubbly names to layouts implemented by Nitro.
func NormalizeLayout(layout string) string {
	switch layout {
	case "club_buy", "loyalty_vip_buy":
		return VIPBuyLayout
	case "club_gift":
		return ClubGiftsLayout
	case "collectibles":
		return DefaultLayout
	case "frontpage", "frontpage_featured", "info_duckets", "info_pets", "info_rentables", "recycler", "recycler_info":
		return InformationLayout
	case "guild_furni":
		return GroupCustomFurnitureLayout
	case "guilds":
		return GroupFrontpageLayout
	case "petcustomization":
		return DefaultLayout
	case "plasto":
		return ColorGroupingLayout
	case "productpage1":
		return SingleBundleLayout
	case "recycler_prizes":
		return DefaultLayout
	case "spaces":
		return SpacesLayout
	default:
		return layout
	}
}
