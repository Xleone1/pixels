package projectile

import (
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// launch reserves one bounded flight slot for a valid selected furniture item.
func (service *Service) launch(active *roomlive.Room, node *configuration.Node, itemID int64, riderID int64) bool {
	direction, distance, pulses, valid := service.parameters(node)
	if !valid {
		return false
	}
	item, found := active.FurnitureItem(itemID)
	if !found {
		return false
	}
	riderOrigin := itemPosition(item)
	if riderID > 0 {
		rider, present := active.UnitMotion(riderID)
		if !present {
			riderID = 0
		} else {
			riderOrigin = rider.Position
		}
	}
	value, _ := service.states.LoadOrStore(active.ID(), &roomState{
		flights: make([]flight, service.maximum), history: make([]snapshot, service.maximum),
	})
	state := value.(*roomState)
	state.mutex.Lock()
	defer state.mutex.Unlock()
	if state.active >= len(state.flights) {
		return false
	}
	for index := range state.flights {
		if state.flights[index].used && state.flights[index].current.ID == itemID {
			return false
		}
	}
	for index := range state.flights {
		if state.flights[index].used {
			continue
		}
		state.flights[index] = flight{
			used: true, roomID: active.ID(), ownerID: item.OwnerPlayerID, riderID: riderID,
			riderOrigin: riderOrigin,
			direction:   direction, remaining: distance, pulsesPerStep: pulses,
			maximumAge: distance * pulses, origin: item, current: item,
			snapshot: snapshot{itemID: item.ID, position: item.Point, altitude: item.Z, moving: true, canMoveFreely: true},
		}
		state.active++
		return true
	}
	return false
}

// parameters validates editor values against the same environment bounds used at save time.
func (service *Service) parameters(node *configuration.Node) (uint8, int, int, bool) {
	if node == nil || len(node.Parameters.Values) != 3 {
		return 0, 0, 0, false
	}
	direction := int(node.Parameters.Values[0])
	distance := int(node.Parameters.Values[1])
	pulses := int(node.Parameters.Values[2])
	if direction < 0 || direction > 7 || distance < 1 || distance > service.maximumDistance ||
		pulses < 1 || pulses > 20 || int64(distance)*int64(pulses) > int64(service.maximumDuration) {
		return 0, 0, 0, false
	}
	return uint8(direction), distance, pulses, true
}

// itemPosition returns the current furniture base as a room position.
func itemPosition(item worldfurniture.Item) worldpath.Position {
	return worldpath.Position{Point: item.Point, Z: item.Z}
}

// remember retains one bounded final snapshot for system-variable inspection.
func (service *Service) remember(state *roomState, value snapshot) {
	value.moving = false
	for index := range state.history {
		if state.history[index].itemID == value.itemID {
			state.history[index] = value
			return
		}
	}
	state.history[state.historyNext%len(state.history)] = value
	state.historyNext = (state.historyNext + 1) % len(state.history)
}

// snapshotValue maps one snapshot property into the shared variable model.
func snapshotValue(roomID int64, current snapshot, name string) (variable.Value, bool) {
	value := int64(0)
	switch name {
	case systemTilesTravelled:
		value = current.tilesTravelled
	case systemUserCollisions:
		value = current.userCollisions
	case systemFurniCollisions:
		value = current.furniCollisions
	case systemPositionX:
		value = int64(current.position.X)
	case systemPositionY:
		value = int64(current.position.Y)
	case systemAltitude:
		value = int64(current.altitude) * 25
	case systemMoving:
		value = boolValue(current.moving)
	case systemCanMoveFreely:
		value = boolValue(current.canMoveFreely)
	default:
		return variable.Value{}, false
	}
	return systemValue(roomID, current.itemID, name, value), true
}

// boolValue maps a runtime flag into a WIRED numeric value.
func boolValue(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
