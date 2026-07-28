package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

const (
	// CurrencyGrantName identifies a cancellable currency mutation.
	CurrencyGrantName = "currency.grant"
	// CurrencyChangedName identifies a committed currency notification.
	CurrencyChangedName = "inventory.currency_changed"
)

// CurrencyGrant fires after validation and before a signed balance mutation.
type CurrencyGrant struct {
	// Player stores the affected immutable player snapshot.
	Player sdkplayer.Player
	// CurrencyType identifies the protocol currency.
	CurrencyType int32
	// Amount stores the signed delta and may be replaced by a listener.
	Amount int64
	// ActorKind identifies the source family of the mutation.
	ActorKind string
	// cancelled stores the current veto state.
	cancelled bool
}

// NewCurrencyGrant creates a cancellable currency mutation event.
func NewCurrencyGrant(player sdkplayer.Player, currencyType int32, amount int64, actorKind string) *CurrencyGrant {
	return &CurrencyGrant{Player: player, CurrencyType: currencyType, Amount: amount, ActorKind: actorKind}
}

// Name returns the stable currency grant event identifier.
func (*CurrencyGrant) Name() string { return CurrencyGrantName }

// Cancelled reports whether the balance mutation was vetoed.
func (event *CurrencyGrant) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the balance mutation is vetoed.
func (event *CurrencyGrant) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned currency event.
func (event *CurrencyGrant) Clone() Mutable {
	cloned := NewCurrencyGrant(event.Player, event.CurrencyType, event.Amount, event.ActorKind)
	cloned.SetCancelled(event.Cancelled())
	return cloned
}

// Apply copies this callback result onto the original currency event.
func (event *CurrencyGrant) Apply(original Mutable) {
	target, ok := original.(*CurrencyGrant)
	if !ok {
		return
	}
	target.Amount = event.Amount
	target.SetCancelled(event.Cancelled())
}

// CurrencyChanged fires after a currency mutation commits.
type CurrencyChanged struct {
	// Player stores the affected immutable player snapshot.
	Player sdkplayer.Player
	// CurrencyType identifies the protocol currency.
	CurrencyType int32
	// Amount stores the resulting absolute balance.
	Amount int64
	// Delta stores the signed committed change.
	Delta int64
	// ActorKind identifies the source family of the mutation.
	ActorKind string
}

// Name returns the stable currency changed event identifier.
func (*CurrencyChanged) Name() string { return CurrencyChangedName }

// CloneEvent returns an isolated callback-owned notification.
func (event *CurrencyChanged) CloneEvent() Event {
	cloned := *event
	return &cloned
}

const (
	// FurniturePlaceName identifies a cancellable furniture placement.
	FurniturePlaceName = "furniture.place"
	// CatalogPurchaseName identifies a cancellable catalog purchase.
	CatalogPurchaseName = "catalog.purchase"
)

// FurniturePlace fires before an inventory item is persisted in a room.
type FurniturePlace struct {
	// Player stores the placement actor.
	Player sdkplayer.Player
	// ItemID identifies the furniture instance.
	ItemID int64
	// RoomID identifies the destination room.
	RoomID int64
	// X stores the mutable floor x coordinate.
	X int
	// Y stores the mutable floor y coordinate.
	Y int
	// Z stores the mutable floor height.
	Z float64
	// Rotation stores the mutable protocol rotation.
	Rotation int
	// WallPosition stores the mutable Nitro wall position.
	WallPosition string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable furniture placement event identifier.
func (*FurniturePlace) Name() string { return FurniturePlaceName }

// Cancelled reports whether placement was vetoed.
func (event *FurniturePlace) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether placement is vetoed.
func (event *FurniturePlace) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned furniture placement event.
func (event *FurniturePlace) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies this callback result onto the original placement event.
func (event *FurniturePlace) Apply(original Mutable) {
	target, ok := original.(*FurniturePlace)
	if !ok {
		return
	}
	target.X, target.Y, target.Z = event.X, event.Y, event.Z
	target.Rotation, target.WallPosition = event.Rotation, event.WallPosition
	target.SetCancelled(event.Cancelled())
}

// CatalogPurchase fires after offer visibility validation and before charging.
type CatalogPurchase struct {
	// Player stores the buyer.
	Player sdkplayer.Player
	// CatalogItemID identifies the offer.
	CatalogItemID int64
	// Amount stores the requested offer quantity.
	Amount int32
	// CreditsCost stores the mutable per-unit credits price.
	CreditsCost int64
	// PointsCost stores the mutable per-unit points price.
	PointsCost int64
	// PointsType stores the mutable points currency type.
	PointsType int32
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable catalog purchase event identifier.
func (*CatalogPurchase) Name() string { return CatalogPurchaseName }

// Cancelled reports whether purchase was vetoed.
func (event *CatalogPurchase) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether purchase is vetoed.
func (event *CatalogPurchase) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned catalog purchase event.
func (event *CatalogPurchase) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies this callback result onto the original purchase event.
func (event *CatalogPurchase) Apply(original Mutable) {
	target, ok := original.(*CatalogPurchase)
	if !ok {
		return
	}
	target.CreditsCost, target.PointsCost, target.PointsType = event.CreditsCost, event.PointsCost, event.PointsType
	target.SetCancelled(event.Cancelled())
}
