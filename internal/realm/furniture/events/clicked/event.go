// Package clicked contains the validated furniture click event.
package clicked

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture click event.
const Name bus.Name = "furniture.clicked"

// Payload describes a validated furniture interaction request.
type Payload struct {
	// PlayerID identifies the actor using the item.
	PlayerID int64
	// ItemID identifies the clicked furniture item.
	ItemID int64
	// RoomID identifies the room containing the item.
	RoomID int64
}
