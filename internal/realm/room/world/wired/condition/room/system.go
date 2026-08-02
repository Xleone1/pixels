package room

import (
	"math"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredvariable "github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

var (
	// roomSystemNames stores inspectable room variable names.
	roomSystemNames = [...]string{"@id", "@owner_id", "@user_count", "@max_users", "@dimensions.x", "@dimensions.y"}
	// furniSystemNames stores inspectable furniture variable names.
	furniSystemNames = [...]string{"@id", "@owner_id", "@sprite_id", "@x", "@y", "@z", "@direction", "@dimensions.x", "@dimensions.y", "@dimensions.z", "@is_stackable", "@can_stand_on", "@can_sit_on", "@can_lay_on", "@state"}
	// userSystemNames stores inspectable user variable names.
	userSystemNames = [...]string{"@id", "@x", "@y", "@z", "@direction", "@handitem", "@effect", "@is_idle", "@username"}
)

// ResolveSystem returns one immutable live room variable.
func (provider *Provider) ResolveSystem(roomID int64, scope wiredvariable.Scope, scopeID int64, name string) (wiredvariable.Value, bool) {
	active, found := provider.rooms.Find(roomID)
	if !found {
		return wiredvariable.Value{}, false
	}
	switch scope {
	case wiredvariable.ScopeRoom:
		return resolveRoomSystem(active, scopeID, name)
	case wiredvariable.ScopeFurni:
		return resolveFurniSystem(active, scopeID, name)
	case wiredvariable.ScopeUser:
		return resolveUserSystem(active, scopeID, name)
	default:
		return wiredvariable.Value{}, false
	}
}

// ListSystem returns every immutable variable for one live target.
func (provider *Provider) ListSystem(roomID int64, scope wiredvariable.Scope, scopeID int64) []wiredvariable.Value {
	names := systemNames(scope)
	result := make([]wiredvariable.Value, 0, len(names))
	for _, name := range names {
		if value, found := provider.ResolveSystem(roomID, scope, scopeID, name); found {
			result = append(result, value)
		}
	}
	return result
}

// systemNames returns the immutable descriptor names for one scope.
func systemNames(scope wiredvariable.Scope) []string {
	switch scope {
	case wiredvariable.ScopeRoom:
		return roomSystemNames[:]
	case wiredvariable.ScopeFurni:
		return furniSystemNames[:]
	case wiredvariable.ScopeUser:
		return userSystemNames[:]
	default:
		return nil
	}
}

// resolveRoomSystem resolves one room metadata value.
func resolveRoomSystem(active *roomlive.Room, scopeID int64, name string) (wiredvariable.Value, bool) {
	snapshot := active.Snapshot()
	if scopeID != snapshot.ID {
		return wiredvariable.Value{}, false
	}
	value := systemValue(snapshot.ID, wiredvariable.ScopeRoom, scopeID, name)
	switch name {
	case "@id":
		value.IntValue = snapshot.ID
	case "@owner_id":
		value.IntValue = snapshot.OwnerPlayerID
	case "@user_count":
		value.IntValue = int64(active.Occupancy().Count)
	case "@max_users":
		value.IntValue = int64(snapshot.MaxUsers)
	case "@dimensions.x", "@dimensions.y":
		width, height, found := active.WorldDimensions()
		if !found {
			return wiredvariable.Value{}, false
		}
		if name == "@dimensions.x" {
			value.IntValue = int64(width)
		} else {
			value.IntValue = int64(height)
		}
	default:
		return wiredvariable.Value{}, false
	}
	return value, true
}

// resolveFurniSystem resolves one placed furniture value.
func resolveFurniSystem(active *roomlive.Room, scopeID int64, name string) (wiredvariable.Value, bool) {
	item, found := active.FurnitureItem(scopeID)
	if !found {
		return wiredvariable.Value{}, false
	}
	value := systemValue(active.ID(), wiredvariable.ScopeFurni, scopeID, name)
	switch name {
	case "@id":
		value.IntValue = item.ID
	case "@owner_id":
		value.IntValue = item.OwnerPlayerID
	case "@sprite_id":
		value.IntValue = int64(item.Definition.SpriteID)
	case "@x":
		value.IntValue = int64(item.Point.X)
	case "@y":
		value.IntValue = int64(item.Point.Y)
	case "@z":
		value.IntValue = hundredths(item.Z.Units())
	case "@direction":
		value.IntValue = int64(item.Rotation)
	case "@dimensions.x":
		value.IntValue = int64(item.Definition.Width)
	case "@dimensions.y":
		value.IntValue = int64(item.Definition.Length)
	case "@dimensions.z":
		value.IntValue = hundredths(item.Definition.StackHeight.Units())
	case "@is_stackable":
		value.IntValue = booleanValue(item.Definition.AllowStack)
	case "@can_stand_on":
		value.IntValue = booleanValue(item.Definition.AllowWalk)
	case "@can_sit_on":
		value.IntValue = booleanValue(item.Definition.AllowSit)
	case "@can_lay_on":
		value.IntValue = booleanValue(item.Definition.AllowLay)
	case "@state":
		value.StringValue = item.ExtraData
	default:
		return wiredvariable.Value{}, false
	}
	return value, true
}

// resolveUserSystem resolves one active room user value.
func resolveUserSystem(active *roomlive.Room, scopeID int64, name string) (wiredvariable.Value, bool) {
	unit, found := active.Unit(scopeID)
	if !found {
		return wiredvariable.Value{}, false
	}
	value := systemValue(active.ID(), wiredvariable.ScopeUser, scopeID, name)
	switch name {
	case "@id":
		value.IntValue = scopeID
	case "@x":
		value.IntValue = int64(unit.Position.Point.X)
	case "@y":
		value.IntValue = int64(unit.Position.Point.Y)
	case "@z":
		value.IntValue = hundredths(unit.Position.Z.Units())
	case "@direction":
		value.IntValue = int64(unit.BodyRotation)
	case "@handitem":
		value.IntValue = int64(unit.HandItem)
	case "@effect":
		value.IntValue = int64(unit.ActiveEffectID)
	case "@is_idle":
		value.IntValue = booleanValue(unit.Idle)
	case "@username":
		value.StringValue = username(active, scopeID)
	default:
		return wiredvariable.Value{}, false
	}
	return value, true
}

// username returns one active occupant name.
func username(active *roomlive.Room, playerID int64) string {
	for _, presence := range active.Presences() {
		if presence.Occupant.PlayerID == playerID {
			return presence.Occupant.Username
		}
	}
	return ""
}

// systemValue creates one immutable variable projection.
func systemValue(roomID int64, scope wiredvariable.Scope, scopeID int64, name string) wiredvariable.Value {
	return wiredvariable.Value{RoomID: roomID, Scope: scope, ScopeID: scopeID, Name: name}
}

// hundredths converts room height units into the Creator Tools integer scale.
func hundredths(value float64) int64 {
	return int64(math.Round(value * 100))
}

// booleanValue converts one domain flag into a WIRED integer.
func booleanValue(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
