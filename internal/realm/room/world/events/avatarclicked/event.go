// Package avatarclicked defines accepted room avatar click events.
package avatarclicked

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies an accepted avatar click.
const Name bus.Name = "room.avatar.clicked"

// Payload describes a validated avatar click.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// PlayerID identifies the clicking player.
	PlayerID int64
	// TargetPlayerID identifies the clicked player.
	TargetPlayerID int64
}
