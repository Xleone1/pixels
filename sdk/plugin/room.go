package plugin

import sdkevent "github.com/niflaot/pixels/sdk/event"

// RoomUpdateParams aliases the plugin-visible partial room mutation.
type RoomUpdateParams = sdkevent.RoomUpdateParams

// RoomSnapshot stores immutable durable room metadata.
type RoomSnapshot struct {
	// ID identifies the room.
	ID int64
	// Version stores the optimistic locking revision.
	Version int64
	// OwnerPlayerID identifies the room owner.
	OwnerPlayerID int64
	// OwnerName stores the owner display name.
	OwnerName string
	// Name stores the visible room name.
	Name string
	// Description stores the visible room description.
	Description string
	// ModelName identifies the room layout.
	ModelName string
	// DoorMode stores the room entry policy.
	DoorMode int16
	// MaxUsers stores the room capacity.
	MaxUsers int
	// Score stores the navigator score.
	Score int
	// CategoryID optionally identifies the navigator category.
	CategoryID *int64
	// TradeMode stores the room trade policy.
	TradeMode int16
	// RollerSpeed stores the room roller cadence.
	RollerSpeed int
	// AllowWalkthrough reports whether units may overlap.
	AllowWalkthrough bool
	// AllowPets reports whether pets may enter.
	AllowPets bool
	// AllowPetsEat reports whether pets may consume food.
	AllowPetsEat bool
	// HideWalls reports whether walls are hidden.
	HideWalls bool
	// HideWired reports whether WIRED boxes are hidden.
	HideWired bool
	// WallThickness stores the rendered wall thickness.
	WallThickness int
	// FloorThickness stores the rendered floor thickness.
	FloorThickness int
	// StaffPicked reports official navigator selection.
	StaffPicked bool
	// PublicRoom reports whether the room is public content.
	PublicRoom bool
}

// RoomAccess exposes bounded room reads and mutations to plugins.
type RoomAccess interface {
	// Update applies a validated partial room mutation.
	Update(int64, RoomUpdateParams) (RoomSnapshot, error)
	// Find returns one immutable durable room snapshot.
	Find(int64) (RoomSnapshot, bool)
	// Occupants lists connected player identifiers inside an active room.
	Occupants(int64) []int64
	// SetMuteAll toggles whole-room muting for an active room.
	SetMuteAll(int64, bool) error
}
