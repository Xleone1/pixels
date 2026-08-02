// Package projectile owns bounded room-cycle custom WIRED projectiles.
package projectile

import (
	"context"
	"sync"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
	"github.com/niflaot/pixels/networking/connection"
)

const (
	// Key identifies the custom projectile extra.
	Key = "wf_xtra_projectile"
	// systemTilesTravelled stores the travelled-tile system variable.
	systemTilesTravelled = "@projectile.animation.tiles_travelled"
	// systemUserCollisions stores the user-collision system variable.
	systemUserCollisions = "@projectile.animation.user_collisions"
	// systemFurniCollisions stores the furniture-collision system variable.
	systemFurniCollisions = "@projectile.animation.furni_collisions"
	// systemPositionX stores the current X system variable.
	systemPositionX = "@projectile.animation.position.x"
	// systemPositionY stores the current Y system variable.
	systemPositionY = "@projectile.animation.position.y"
	// systemAltitude stores the current altitude system variable.
	systemAltitude = "@projectile.animation.altitude"
	// systemMoving stores the movement-state system variable.
	systemMoving = "@projectile.animation.moving"
	// systemCanMoveFreely stores the next-step availability system variable.
	systemCanMoveFreely = "@can_move_freely"
)

var systemNames = [...]string{
	systemTilesTravelled, systemUserCollisions, systemFurniCollisions,
	systemPositionX, systemPositionY, systemAltitude, systemMoving, systemCanMoveFreely,
}

// FurnitureMover adapts the furniture service to final projectile persistence.
type FurnitureMover interface {
	// Move repositions one already placed item.
	Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error)
}

// Service owns bounded projectile state per active room.
type Service struct {
	// rooms resolves active room worlds.
	rooms *roomlive.Registry
	// furniture persists final placements.
	furniture FurnitureMover
	// connections projects normal room packets.
	connections *connection.Registry
	// maximum bounds active flights and retained final snapshots per room.
	maximum int
	// maximumDistance bounds one flight path.
	maximumDistance int
	// maximumDuration bounds one flight lifetime in room pulses.
	maximumDuration int
	// states stores room-owned bounded projectile state.
	states sync.Map
}

// roomState stores fixed-capacity active flights and retained final snapshots.
type roomState struct {
	// mutex serializes room-cycle advancement and variable reads.
	mutex sync.RWMutex
	// flights stores fixed active slots.
	flights []flight
	// history stores fixed retained final snapshots.
	history []snapshot
	// historyNext stores the next retained snapshot slot.
	historyNext int
	// active stores occupied flight slots.
	active int
}

// flight stores one bounded projectile animation.
type flight struct {
	// used reports that this slot is active.
	used bool
	// roomID identifies the owning room.
	roomID int64
	// ownerID identifies the durable furniture owner.
	ownerID int64
	// riderID optionally identifies the visually-following actor.
	riderID int64
	// riderOrigin stores the actor position restored after persistence rollback.
	riderOrigin worldpath.Position
	// direction stores a Habbo rotation from zero through seven.
	direction uint8
	// remaining stores pending tile steps.
	remaining int
	// pulsesPerStep stores the room pulses between steps.
	pulsesPerStep int
	// pulse stores the current step pulse cursor.
	pulse int
	// age stores elapsed room pulses.
	age int
	// maximumAge stores the bounded flight lifetime.
	maximumAge int
	// origin stores the durable starting placement.
	origin worldfurniture.Item
	// current stores the current projected placement.
	current worldfurniture.Item
	// snapshot stores observable flight counters and position.
	snapshot snapshot
}

// snapshot stores retained projectile system-variable values.
type snapshot struct {
	// itemID identifies the projectile furniture.
	itemID int64
	// position stores its last projected origin tile.
	position grid.Point
	// altitude stores its last projected base height.
	altitude grid.Height
	// tilesTravelled stores successful step count.
	tilesTravelled int64
	// userCollisions stores swept user collision count.
	userCollisions int64
	// furniCollisions stores swept furniture collision count.
	furniCollisions int64
	// moving reports whether the animation remains active.
	moving bool
	// canMoveFreely reports whether the next swept step is unobstructed.
	canMoveFreely bool
}

// systemValue creates one immutable furni-scoped system variable.
func systemValue(roomID int64, itemID int64, name string, value int64) variable.Value {
	return variable.Value{RoomID: roomID, Scope: variable.ScopeFurni, ScopeID: itemID, Name: name, IntValue: value}
}
