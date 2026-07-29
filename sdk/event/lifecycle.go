package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

const (
	// FurnitureMoveName identifies a cancellable furniture move.
	FurnitureMoveName = "furniture.move"
	// FurniturePickupName identifies a cancellable furniture pickup.
	FurniturePickupName = "furniture.pickup"
	// RoomCreateName identifies a cancellable room creation.
	RoomCreateName = "room.create"
	// MarketplaceListName identifies a cancellable marketplace listing.
	MarketplaceListName = "marketplace.list"
	// MarketplaceBuyName identifies a cancellable marketplace purchase.
	MarketplaceBuyName = "marketplace.buy"
)

// FurnitureMove fires before a placed furniture item is moved.
type FurnitureMove struct {
	// Player stores the immutable actor snapshot.
	Player sdkplayer.Player
	// ItemID identifies the furniture item being moved.
	ItemID int64
	// RoomID identifies the room containing the item.
	RoomID int64
	// X stores the mutable destination tile coordinate.
	X int
	// Y stores the mutable destination tile coordinate.
	Y int
	// Rotation stores the mutable furniture rotation.
	Rotation int
	// Z stores the mutable destination height.
	Z float64
	// WallPosition stores mutable Nitro wall coordinates.
	WallPosition string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable furniture-move event identifier.
func (*FurnitureMove) Name() string { return FurnitureMoveName }

// Cancelled reports whether the move was vetoed.
func (event *FurnitureMove) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the move is vetoed.
func (event *FurnitureMove) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned furniture move.
func (event *FurnitureMove) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies mutable movement state onto the original event.
func (event *FurnitureMove) Apply(original Mutable) {
	if target, ok := original.(*FurnitureMove); ok {
		target.X, target.Y, target.Z, target.Rotation, target.WallPosition, target.cancelled =
			event.X, event.Y, event.Z, event.Rotation, event.WallPosition, event.cancelled
	}
}

// FurniturePickup fires before a placed furniture item returns to inventory.
type FurniturePickup struct {
	// Player stores the immutable actor snapshot.
	Player sdkplayer.Player
	// ItemID identifies the furniture item being picked up.
	ItemID int64
	// RoomID identifies the room containing the item.
	RoomID int64
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable furniture-pickup event identifier.
func (*FurniturePickup) Name() string { return FurniturePickupName }

// Cancelled reports whether the pickup was vetoed.
func (event *FurniturePickup) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the pickup is vetoed.
func (event *FurniturePickup) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned furniture pickup.
func (event *FurniturePickup) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies the veto state onto the original event.
func (event *FurniturePickup) Apply(original Mutable) {
	if target, ok := original.(*FurniturePickup); ok {
		target.cancelled = event.cancelled
	}
}

// RoomCreate fires after native validation and before room persistence.
type RoomCreate struct {
	// Player stores the immutable room owner snapshot.
	Player sdkplayer.Player
	// RoomName stores the mutable visible room name.
	RoomName string
	// Description stores the mutable visible description.
	Description string
	// ModelName stores the mutable layout identifier.
	ModelName string
	// MaxUsers stores the mutable room capacity.
	MaxUsers int
	// CategoryID stores the mutable navigator category.
	CategoryID *int64
	// TradeMode stores the mutable room trade policy.
	TradeMode int16
	// Tags stores mutable room tags.
	Tags []string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable room-creation event identifier.
func (*RoomCreate) Name() string { return RoomCreateName }

// Cancelled reports whether room creation was vetoed.
func (event *RoomCreate) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether room creation is vetoed.
func (event *RoomCreate) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned room creation.
func (event *RoomCreate) Clone() Mutable {
	cloned := *event
	cloned.CategoryID = clonePointer(event.CategoryID)
	cloned.Tags = append([]string(nil), event.Tags...)
	return &cloned
}

// Apply copies mutable room creation state onto the original event.
func (event *RoomCreate) Apply(original Mutable) {
	if target, ok := original.(*RoomCreate); ok {
		target.RoomName, target.Description, target.ModelName, target.MaxUsers =
			event.RoomName, event.Description, event.ModelName, event.MaxUsers
		target.CategoryID, target.TradeMode = clonePointer(event.CategoryID), event.TradeMode
		target.Tags, target.cancelled = append([]string(nil), event.Tags...), event.cancelled
	}
}

// MarketplaceList fires before a marketplace listing transaction.
type MarketplaceList struct {
	// Player stores the immutable seller snapshot.
	Player sdkplayer.Player
	// ItemID identifies the furniture item being listed.
	ItemID int64
	// RawPrice stores the mutable seller price before commission.
	RawPrice int64
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable marketplace-list event identifier.
func (*MarketplaceList) Name() string { return MarketplaceListName }

// Cancelled reports whether the listing was vetoed.
func (event *MarketplaceList) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the listing is vetoed.
func (event *MarketplaceList) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned marketplace listing.
func (event *MarketplaceList) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies mutable listing state onto the original event.
func (event *MarketplaceList) Apply(original Mutable) {
	if target, ok := original.(*MarketplaceList); ok {
		target.RawPrice, target.cancelled = event.RawPrice, event.cancelled
	}
}

// MarketplaceBuy fires before a marketplace buyer is charged.
type MarketplaceBuy struct {
	// Buyer stores the immutable buyer snapshot.
	Buyer sdkplayer.Player
	// ListingID identifies the marketplace listing.
	ListingID int64
	// SellerID identifies the listing owner.
	SellerID int64
	// BuyerPrice stores the mutable final charged price.
	BuyerPrice int64
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable marketplace-buy event identifier.
func (*MarketplaceBuy) Name() string { return MarketplaceBuyName }

// Cancelled reports whether the purchase was vetoed.
func (event *MarketplaceBuy) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the purchase is vetoed.
func (event *MarketplaceBuy) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned marketplace purchase.
func (event *MarketplaceBuy) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies mutable purchase state onto the original event.
func (event *MarketplaceBuy) Apply(original Mutable) {
	if target, ok := original.(*MarketplaceBuy); ok {
		target.BuyerPrice, target.cancelled = event.BuyerPrice, event.cancelled
	}
}
