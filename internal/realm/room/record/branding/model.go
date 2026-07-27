// Package branding coordinates durable room branding projected through compatible furniture.
package branding

import "time"

// Kind identifies the Nitro branding furniture behavior.
type Kind string

const (
	// KindBackground identifies a non-clickable room background.
	KindBackground Kind = "background"
	// KindBillboard identifies a clickable room billboard.
	KindBillboard Kind = "billboard"
)

// Valid reports whether the branding kind is supported.
func (kind Kind) Valid() bool {
	return kind == KindBackground || kind == KindBillboard
}

// Config contains one durable room branding configuration.
type Config struct {
	// ID identifies the branding configuration.
	ID int64 `json:"id"`
	// RoomID identifies the containing room.
	RoomID int64 `json:"roomId"`
	// FurnitureItemID identifies the projected furniture item.
	FurnitureItemID int64 `json:"furnitureItemId"`
	// Kind identifies background or billboard behavior.
	Kind Kind `json:"kind"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef"`
	// ImageURL stores the durable public image URL.
	ImageURL string `json:"imageUrl"`
	// ClickURL stores the optional billboard target URL.
	ClickURL string `json:"clickUrl"`
	// State stores the client visual state.
	State int16 `json:"state"`
	// OffsetX stores the horizontal renderer offset.
	OffsetX int `json:"offsetX"`
	// OffsetY stores the vertical renderer offset.
	OffsetY int `json:"offsetY"`
	// OffsetZ stores the depth renderer offset.
	OffsetZ int `json:"offsetZ"`
	// Enabled reports whether branding is projected.
	Enabled bool `json:"enabled"`
	// CreatedByPlayerID identifies the creating actor.
	CreatedByPlayerID int64 `json:"createdByPlayerId"`
	// UpdatedByPlayerID identifies the latest actor.
	UpdatedByPlayerID int64 `json:"updatedByPlayerId"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores the latest update time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores optimistic mutation order.
	Version int64 `json:"version"`
}

// CompatibleItem describes one placed branding-capable furniture item.
type CompatibleItem struct {
	// ItemID identifies the furniture instance.
	ItemID int64 `json:"itemId"`
	// DefinitionID identifies the furniture definition.
	DefinitionID int64 `json:"definitionId"`
	// Name stores the stable furniture class name.
	Name string `json:"name"`
	// PublicName stores the visible furniture name.
	PublicName string `json:"publicName"`
	// Kind identifies the supported branding behavior.
	Kind Kind `json:"kind"`
	// Configured reports whether the item has durable branding.
	Configured bool `json:"configured"`
}

// Projection contains one committed branding mutation and furniture projection data.
type Projection struct {
	// Config contains the committed branding record.
	Config Config
	// SpriteID identifies the client furniture sprite.
	SpriteID int
	// OwnerPlayerID identifies the furniture owner.
	OwnerPlayerID int64
	// X stores floor position.
	X int
	// Y stores floor position.
	Y int
	// Z stores floor height.
	Z float64
	// Rotation stores furniture direction.
	Rotation int
	// ExtraData stores the canonical Nitro map source.
	ExtraData string
	// InteractionType stores the server projection behavior.
	InteractionType string
}

// Mutation contains validated branding mutation input.
type Mutation struct {
	// RoomID identifies the containing room.
	RoomID int64
	// FurnitureItemID identifies the target furniture.
	FurnitureItemID int64
	// Kind identifies background or billboard behavior.
	Kind Kind
	// AssetRef stores an opaque CMS-owned reference.
	AssetRef string
	// ImageURL stores a public durable URL.
	ImageURL string
	// ClickURL stores an optional public target.
	ClickURL string
	// State stores the visual state.
	State int16
	// OffsetX stores the horizontal renderer offset.
	OffsetX int
	// OffsetY stores the vertical renderer offset.
	OffsetY int
	// OffsetZ stores the depth renderer offset.
	OffsetZ int
	// ExpectedVersion stores zero for create or the current version for update.
	ExpectedVersion int64
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
	// Reason explains the mutation.
	Reason string
}
