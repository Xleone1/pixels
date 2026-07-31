// Package acted defines accepted room user action events.
package acted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies an accepted user action.
const Name bus.Name = "room.unit.acted"

// Payload describes one action using the WIRED action discriminator.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// PlayerID identifies the acting player.
	PlayerID int64
	// Action stores the WIRED user action type.
	Action int32
	// Value stores an optional sign, dance, or expression variant.
	Value int32
}
