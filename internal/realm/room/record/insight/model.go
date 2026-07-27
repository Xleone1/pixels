// Package insight provides durable and live room observability.
package insight

import (
	"context"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// Stats contains durable room content counters with optional live occupancy.
type Stats struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// Furniture stores every placed furniture item.
	Furniture int `json:"furniture"`
	// FloorFurniture stores placed floor furniture.
	FloorFurniture int `json:"floorFurniture"`
	// WallFurniture stores placed wall furniture.
	WallFurniture int `json:"wallFurniture"`
	// WiredFurniture stores placed WIRED furniture without exposing configuration.
	WiredFurniture int `json:"wiredFurniture"`
	// Bots stores placed bots.
	Bots int `json:"bots"`
	// Pets stores placed pets.
	Pets int `json:"pets"`
	// Branding stores enabled room branding configurations.
	Branding int `json:"branding"`
	// Occupants stores current live occupancy.
	Occupants int `json:"occupants"`
	// MaxUsers stores room capacity.
	MaxUsers int `json:"maxUsers"`
	// RuntimeAvailable reports whether live state is loaded.
	RuntimeAvailable bool `json:"runtimeAvailable"`
}

// Store reads durable room observability counters.
type Store interface {
	// Stats reads durable counters for one room.
	Stats(context.Context, int64) (Stats, bool, error)
}

// Service combines durable counters with active runtime state.
type Service struct {
	// store reads durable room counters.
	store Store
	// runtime resolves loaded rooms.
	runtime *roomlive.Registry
}

// New creates a room insight service.
func New(store Store, runtime *roomlive.Registry) *Service {
	return &Service{store: store, runtime: runtime}
}

// Stats reads durable counters and overlays live occupancy.
func (service *Service) Stats(ctx context.Context, roomID int64) (Stats, bool, error) {
	stats, found, err := service.store.Stats(ctx, roomID)
	if err != nil || !found {
		return stats, found, err
	}
	if active, activeFound := service.runtime.Find(roomID); activeFound {
		stats.RuntimeAvailable = true
		stats.Occupants = active.Occupancy().Count
	}
	return stats, true, nil
}

// Profile reads a bounded active-room profile.
func (service *Service) Profile(roomID int64) roomlive.Profile {
	active, found := service.runtime.Find(roomID)
	if !found {
		return roomlive.Profile{}
	}
	return active.Profile()
}
