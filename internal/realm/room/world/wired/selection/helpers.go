package selection

import (
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// idSet deduplicates the common small selection without heap allocation.
type idSet struct {
	// inline stores the first unique identifiers.
	inline [8]int64
	// count stores populated inline identifiers.
	count int
	// overflow indexes selections beyond the inline capacity.
	overflow map[int64]struct{}
}

// add reports whether a positive identifier was inserted.
func (set *idSet) add(value int64) bool {
	if set.overflow != nil {
		if _, found := set.overflow[value]; found {
			return false
		}
		set.overflow[value] = struct{}{}
		return true
	}
	for _, current := range set.inline[:set.count] {
		if current == value {
			return false
		}
	}
	if set.count < len(set.inline) {
		set.inline[set.count] = value
		set.count++
		return true
	}
	set.overflow = make(map[int64]struct{}, set.count+1)
	for _, current := range set.inline[:set.count] {
		set.overflow[current] = struct{}{}
	}
	set.overflow[value] = struct{}{}
	return true
}

// appendFurni appends unique positive furniture targets up to a hard limit.
func appendFurni(selection *Selection, seen *idSet, values []record.Target, limit int) {
	for _, value := range values {
		if value.ItemID <= 0 || len(selection.Furni) >= limit {
			continue
		}
		if !seen.add(value.ItemID) {
			continue
		}
		selection.Furni = append(selection.Furni, value)
	}
}

// appendActors appends unique positive actor identifiers up to a hard limit.
func appendActors(selection *Selection, seen *idSet, values []int64, limit int) {
	for _, value := range values {
		if value == 0 || len(selection.Actors) >= limit {
			continue
		}
		if !seen.add(value) {
			continue
		}
		selection.Actors = append(selection.Actors, value)
	}
}

// valueAt returns one parameter or zero.
func valueAt(values []int32, index int) int32 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}

// area normalizes two inclusive corners.
func area(values []int32) (int, int, int, int) {
	x1, y1, x2, y2 := int(valueAt(values, 0)), int(valueAt(values, 1)), int(valueAt(values, 2)), int(valueAt(values, 3))
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return x1, y1, x2, y2
}
