package room

import (
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredvariable "github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// FurniByType returns furniture sharing any selected target sprite.
func (view *View) FurniByType(targets []record.Target) []record.Target {
	sprites := make(map[int32]struct{}, len(targets))
	for _, target := range targets {
		sprites[target.SpriteID] = struct{}{}
	}
	result := make([]record.Target, 0)
	for _, item := range view.active.FurnitureItems() {
		if _, found := sprites[int32(item.Definition.SpriteID)]; found {
			result = append(result, target(item.ID, int32(item.Definition.SpriteID)))
		}
	}
	return result
}

// FurniByAltitude returns furniture between inclusive hundredth-unit bounds.
func (view *View) FurniByAltitude(minimum int32, maximum int32) []record.Target {
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	result := make([]record.Target, 0)
	for _, item := range view.active.FurnitureItems() {
		height := int32(item.Z.Units() * 100)
		if height >= minimum && height <= maximum {
			result = append(result, target(item.ID, int32(item.Definition.SpriteID)))
		}
	}
	return result
}

// FurniOnFurniture returns items physically stacked on selected bases.
func (view *View) FurniOnFurniture(targets []record.Target) []record.Target {
	items := view.active.FurnitureItems()
	result := make([]record.Target, 0)
	for _, candidate := range items {
		for _, selected := range targets {
			base, found := view.active.FurnitureItem(selected.ItemID)
			if found && candidate.ID != base.ID && candidate.Z >= base.Top() && footprintsOverlap(base, candidate) {
				result = append(result, target(candidate.ID, int32(candidate.Definition.SpriteID)))
				break
			}
		}
	}
	return result
}

// FurniInArea returns furniture origins inside an inclusive rectangle.
func (view *View) FurniInArea(x1 int, y1 int, x2 int, y2 int) []record.Target {
	result := make([]record.Target, 0)
	for _, item := range view.active.FurnitureItems() {
		x, y := int(item.Point.X), int(item.Point.Y)
		if x >= x1 && x <= x2 && y >= y1 && y <= y2 {
			result = append(result, target(item.ID, int32(item.Definition.SpriteID)))
		}
	}
	return result
}

// FurniByVariable returns furniture with one currently loaded variable.
func (view *View) FurniByVariable(name string) []record.Target {
	if view.variables == nil || name == "" {
		return nil
	}
	result := make([]record.Target, 0)
	for _, item := range view.active.FurnitureItems() {
		if view.variables.Exists(view.active.ID(), wiredvariable.ScopeFurni, item.ID, name) {
			result = append(result, target(item.ID, int32(item.Definition.SpriteID)))
		}
	}
	return result
}

// UsersByVariable returns actors with one currently loaded variable.
func (view *View) UsersByVariable(name string) []int64 {
	if view.variables == nil || name == "" {
		return nil
	}
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		scopeID := unit.PlayerID
		if scopeID == 0 {
			scopeID = unit.EntityKey
		}
		if view.variables.Exists(view.active.ID(), wiredvariable.ScopeUser, scopeID, name) {
			result = append(result, unit.EntityKey)
		}
	}
	return result
}

// target creates a minimal dynamic target record.
func target(itemID int64, spriteID int32) record.Target {
	return record.Target{ItemID: itemID, SpriteID: spriteID}
}
