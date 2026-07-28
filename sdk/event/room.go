package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

const (
	// RoomUpdateName identifies a cancellable room settings mutation.
	RoomUpdateName = "room.update"
	// RoomUnitMoveName identifies a cancellable unit movement attempt.
	RoomUnitMoveName = "room.unit.move"
	// RoomEnterAttemptName identifies a cancellable room admission attempt.
	RoomEnterAttemptName = "room.enter.attempt"
)

// RoomEnterAttempt fires after native access checks and before room mutation.
type RoomEnterAttempt struct {
	// Player stores the entering player.
	Player sdkplayer.Player
	// RoomID identifies the destination room.
	RoomID int64
	// Trusted reports a server-controlled admission.
	Trusted bool
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable room admission event identifier.
func (*RoomEnterAttempt) Name() string { return RoomEnterAttemptName }

// Cancelled reports whether room admission was vetoed.
func (event *RoomEnterAttempt) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether room admission is vetoed.
func (event *RoomEnterAttempt) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned room admission event.
func (event *RoomEnterAttempt) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies this callback result onto the original room admission event.
func (event *RoomEnterAttempt) Apply(original Mutable) {
	target, ok := original.(*RoomEnterAttempt)
	if ok {
		target.SetCancelled(event.Cancelled())
	}
}

// RoomUpdateParams contains plugin-visible optional room settings.
type RoomUpdateParams struct {
	// Name replaces the visible room name.
	Name *string
	// Description replaces the visible room description.
	Description *string
	// CategoryID replaces or clears the navigator category.
	CategoryID **int64
	// Tags replaces the normalized tag set.
	Tags *[]string
	// MaxUsers replaces room capacity.
	MaxUsers *int
	// DoorMode replaces room access mode.
	DoorMode *int16
	// Password contains a new plaintext room password.
	Password *string
	// TradeMode replaces room trading behavior.
	TradeMode *int16
	// RollerSpeed replaces roller cadence.
	RollerSpeed *int
	// AllowWalkthrough replaces unit walkthrough behavior.
	AllowWalkthrough *bool
	// AllowPets replaces pet admission behavior.
	AllowPets *bool
	// AllowPetsEat replaces pet food behavior.
	AllowPetsEat *bool
	// HideWalls replaces wall visibility.
	HideWalls *bool
	// HideWired replaces WIRED box visibility.
	HideWired *bool
	// WallThickness replaces wall thickness.
	WallThickness *int
	// FloorThickness replaces floor thickness.
	FloorThickness *int
	// ChatMode replaces room chat mode.
	ChatMode *int16
	// ChatWeight replaces room chat weight.
	ChatWeight *int16
	// ChatSpeed replaces room chat speed.
	ChatSpeed *int16
	// ChatDistance replaces room chat distance.
	ChatDistance *int16
	// ChatProtection replaces flood protection.
	ChatProtection *int16
	// ModerationMute replaces mute policy.
	ModerationMute *int16
	// ModerationKick replaces kick policy.
	ModerationKick *int16
	// ModerationBan replaces ban policy.
	ModerationBan *int16
	// StaffPicked replaces official selection state.
	StaffPicked *bool
}

// RoomUpdate fires after native validation and before persistence.
type RoomUpdate struct {
	// Player stores the actor when the mutation has one.
	Player sdkplayer.Player
	// RoomID identifies the affected room.
	RoomID int64
	// Params stores the partial mutation and may be replaced by listeners.
	Params RoomUpdateParams
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable room update event identifier.
func (*RoomUpdate) Name() string { return RoomUpdateName }

// Cancelled reports whether persistence was vetoed.
func (event *RoomUpdate) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether persistence is vetoed.
func (event *RoomUpdate) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned room update event.
func (event *RoomUpdate) Clone() Mutable {
	cloned := *event
	cloned.Params = cloneRoomUpdateParams(event.Params)
	return &cloned
}

// Apply copies this callback result onto the original room update event.
func (event *RoomUpdate) Apply(original Mutable) {
	target, ok := original.(*RoomUpdate)
	if !ok {
		return
	}
	target.Params = cloneRoomUpdateParams(event.Params)
	target.SetCancelled(event.Cancelled())
}

// RoomUnitMove fires before a live unit accepts a path target.
type RoomUnitMove struct {
	// Player stores the moving player snapshot.
	Player sdkplayer.Player
	// RoomID identifies the active room.
	RoomID int64
	// TargetX stores the destination x coordinate.
	TargetX int
	// TargetY stores the destination y coordinate.
	TargetY int
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable movement event identifier.
func (*RoomUnitMove) Name() string { return RoomUnitMoveName }

// Cancelled reports whether movement was vetoed.
func (event *RoomUnitMove) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether movement is vetoed.
func (event *RoomUnitMove) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned movement event.
func (event *RoomUnitMove) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies this callback result onto the original movement event.
func (event *RoomUnitMove) Apply(original Mutable) {
	target, ok := original.(*RoomUnitMove)
	if !ok {
		return
	}
	target.TargetX = event.TargetX
	target.TargetY = event.TargetY
	target.SetCancelled(event.Cancelled())
}

// cloneRoomUpdateParams protects slice and pointer fields between callbacks.
func cloneRoomUpdateParams(params RoomUpdateParams) RoomUpdateParams {
	return RoomUpdateParams{
		Name:             clonePointer(params.Name),
		Description:      clonePointer(params.Description),
		CategoryID:       cloneOptionalPointer(params.CategoryID),
		Tags:             cloneSlicePointer(params.Tags),
		MaxUsers:         clonePointer(params.MaxUsers),
		DoorMode:         clonePointer(params.DoorMode),
		Password:         clonePointer(params.Password),
		TradeMode:        clonePointer(params.TradeMode),
		RollerSpeed:      clonePointer(params.RollerSpeed),
		AllowWalkthrough: clonePointer(params.AllowWalkthrough),
		AllowPets:        clonePointer(params.AllowPets),
		AllowPetsEat:     clonePointer(params.AllowPetsEat),
		HideWalls:        clonePointer(params.HideWalls),
		HideWired:        clonePointer(params.HideWired),
		WallThickness:    clonePointer(params.WallThickness),
		FloorThickness:   clonePointer(params.FloorThickness),
		ChatMode:         clonePointer(params.ChatMode),
		ChatWeight:       clonePointer(params.ChatWeight),
		ChatSpeed:        clonePointer(params.ChatSpeed),
		ChatDistance:     clonePointer(params.ChatDistance),
		ChatProtection:   clonePointer(params.ChatProtection),
		ModerationMute:   clonePointer(params.ModerationMute),
		ModerationKick:   clonePointer(params.ModerationKick),
		ModerationBan:    clonePointer(params.ModerationBan),
		StaffPicked:      clonePointer(params.StaffPicked),
	}
}

// clonePointer copies one optional scalar value.
func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// cloneOptionalPointer copies a nullable optional scalar value.
func cloneOptionalPointer[T any](value **T) **T {
	if value == nil {
		return nil
	}
	if *value == nil {
		var cleared *T
		return &cleared
	}
	cloned := **value
	pointer := &cloned
	return &pointer
}

// cloneSlicePointer copies one optional slice value.
func cloneSlicePointer[T any](value *[]T) *[]T {
	if value == nil {
		return nil
	}
	cloned := append([]T(nil), (*value)...)
	return &cloned
}
