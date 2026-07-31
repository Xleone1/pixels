// Package selection resolves WIRED dynamic target sets.
package selection

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Selection stores resolved furniture and actor targets in stable order.
type Selection struct {
	// Furni stores resolved furniture records.
	Furni []record.Target
	// Actors stores resolved room entity keys.
	Actors []int64
	// FurniResolved reports that a selector replaced furniture targets, even when empty.
	FurniResolved bool
	// ActorsResolved reports that a selector replaced actor targets, even when empty.
	ActorsResolved bool
}

// View exposes focused live-room queries required by selectors.
type View interface {
	// UsersByAction returns actors currently performing one action.
	UsersByAction(int32) []int64
	// UsersByName returns actors matching an optional exact username.
	UsersByName(string) []int64
	// UsersOnFurniture returns actors occupying selected furniture.
	UsersOnFurniture([]record.Target) []int64
	// UsersInGroup returns actors belonging to the room group.
	UsersInGroup() []int64
	// UsersByHanditem returns actors holding one hand item.
	UsersByHanditem(int32) []int64
	// UsersByType returns actors of one room-unit kind.
	UsersByType(int32) []int64
	// UsersInArea returns actors inside an inclusive rectangle.
	UsersInArea(int, int, int, int) []int64
	// UsersByTeam returns actors assigned to one WIRED game team.
	UsersByTeam(int32) []int64
	// FurniByType returns furniture sharing selected sprite types.
	FurniByType([]record.Target) []record.Target
	// FurniByAltitude returns furniture inside an inclusive fixed-point range.
	FurniByAltitude(int32, int32) []record.Target
	// FurniOnFurniture returns furniture stacked on selected bases.
	FurniOnFurniture([]record.Target) []record.Target
	// FurniInArea returns furniture whose origin lies inside a rectangle.
	FurniInArea(int, int, int, int) []record.Target
	// FurniByVariable returns furniture carrying a named variable.
	FurniByVariable(string) []record.Target
	// UsersByVariable returns actors carrying a named variable.
	UsersByVariable(string) []int64
}

// Provider resolves a selection-capable view for one active room.
type Provider interface {
	// SelectionView returns one live-room selection view.
	SelectionView(int64) (View, bool)
}

// Resolver evaluates selector nodes against immutable configuration and live state.
type Resolver struct {
	// limit bounds each resolved target family.
	limit int
}

// New creates a bounded selection resolver.
func New(limit int) *Resolver {
	if limit <= 0 {
		limit = 1000
	}
	return &Resolver{limit: limit}
}

// ResolveStack evaluates every selector in stable node order.
func (resolver *Resolver) ResolveStack(ctx context.Context, stack *configuration.Stack, event trigger.Event, generation *configuration.Generation, view View) (Selection, error) {
	return resolver.resolve(ctx, stack, event, generation, view, 0)
}
