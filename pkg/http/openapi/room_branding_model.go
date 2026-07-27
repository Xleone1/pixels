package openapi

// RoomBrandingReadRequest contains one attributed branding read.
type RoomBrandingReadRequest struct {
	RoomIDRequest
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `query:"actorPlayerId" required:"true" minimum:"1"`
}

// RoomBrandingMutationRequest documents one attributed branding replacement.
type RoomBrandingMutationRequest struct {
	RoomBrandingItemRequest
	// Kind identifies background or billboard behavior.
	Kind string `json:"kind" required:"true" enum:"background,billboard"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef" maxLength:"255"`
	// ImageURL stores the durable public image URL.
	ImageURL string `json:"imageUrl" required:"true" format:"uri"`
	// ClickURL stores the optional billboard destination.
	ClickURL string `json:"clickUrl,omitempty" format:"uri"`
	// State stores the renderer visual state.
	State int16 `json:"state" minimum:"0" maximum:"255"`
	// OffsetX stores the horizontal renderer offset.
	OffsetX int `json:"offsetX" minimum:"-4096" maximum:"4096"`
	// OffsetY stores the vertical renderer offset.
	OffsetY int `json:"offsetY" minimum:"-4096" maximum:"4096"`
	// OffsetZ stores the depth renderer offset.
	OffsetZ int `json:"offsetZ" minimum:"-4096" maximum:"4096"`
	// ExpectedVersion stores zero for creation or the current version for update.
	ExpectedVersion int64 `json:"expectedVersion" minimum:"0"`
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// RoomBrandingDisableRequest documents one attributed branding disable.
type RoomBrandingDisableRequest struct {
	RoomBrandingIDRequest
	// ExpectedVersion stores the current branding version.
	ExpectedVersion int64 `json:"expectedVersion" required:"true" minimum:"1"`
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// RoomBrandingItemRequest contains room and furniture path identifiers.
type RoomBrandingItemRequest struct {
	RoomIDRequest
	// ItemID identifies placed branding-compatible furniture.
	ItemID int64 `path:"itemId" required:"true" minimum:"1"`
}

// RoomBrandingIDRequest contains room and branding path identifiers.
type RoomBrandingIDRequest struct {
	RoomIDRequest
	// BrandingID identifies one durable branding configuration.
	BrandingID int64 `path:"brandingId" required:"true" minimum:"1"`
}

// RoomBrandingResponse describes one durable branding configuration.
type RoomBrandingResponse struct {
	// ID identifies the branding record.
	ID int64 `json:"id" required:"true"`
	// RoomID identifies the containing room.
	RoomID int64 `json:"roomId" required:"true"`
	// FurnitureItemID identifies placed furniture.
	FurnitureItemID int64 `json:"furnitureItemId" required:"true"`
	// Kind identifies background or billboard behavior.
	Kind string `json:"kind" required:"true"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef"`
	// ImageURL stores the durable public image URL.
	ImageURL string `json:"imageUrl" required:"true"`
	// ClickURL stores the optional billboard destination.
	ClickURL string `json:"clickUrl"`
	// State stores the renderer visual state.
	State int16 `json:"state"`
	// OffsetX stores the horizontal renderer offset.
	OffsetX int `json:"offsetX"`
	// OffsetY stores the vertical renderer offset.
	OffsetY int `json:"offsetY"`
	// OffsetZ stores the depth renderer offset.
	OffsetZ int `json:"offsetZ"`
	// Enabled reports whether Pixels projects this branding.
	Enabled bool `json:"enabled" required:"true"`
	// UpdatedAt stores the latest committed mutation.
	UpdatedAt string `json:"updatedAt" required:"true" format:"date-time"`
	// Version stores optimistic mutation order.
	Version int64 `json:"version" required:"true" minimum:"1"`
}

// RoomBrandingListResponse contains room branding rows.
type RoomBrandingListResponse struct {
	// Total stores returned configuration count.
	Total int `json:"total" required:"true"`
	// Items stores branding configurations.
	Items []RoomBrandingResponse `json:"items" required:"true"`
}

// RoomBrandingFurnitureResponse describes one compatible placed furniture item.
type RoomBrandingFurnitureResponse struct {
	// ItemID identifies placed furniture.
	ItemID int64 `json:"itemId" required:"true"`
	// DefinitionID identifies the furniture definition.
	DefinitionID int64 `json:"definitionId" required:"true"`
	// Name stores the definition internal name.
	Name string `json:"name" required:"true"`
	// PublicName stores the visible definition name.
	PublicName string `json:"publicName" required:"true"`
	// Kind identifies background or billboard behavior.
	Kind string `json:"kind" required:"true"`
	// Configured reports whether the furniture has durable branding.
	Configured bool `json:"configured" required:"true"`
}

// RoomBrandingFurnitureListResponse contains compatible furniture rows.
type RoomBrandingFurnitureListResponse struct {
	// Total stores returned furniture count.
	Total int `json:"total" required:"true"`
	// Items stores compatible placed furniture.
	Items []RoomBrandingFurnitureResponse `json:"items" required:"true"`
}
