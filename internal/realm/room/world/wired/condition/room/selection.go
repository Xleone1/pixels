package room

import (
	"strconv"
	"strings"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
)

// SelectionView returns this adapter as a dynamic selector view.
func (provider *Provider) SelectionView(roomID int64) (selection.View, bool) {
	value, found := provider.View(roomID)
	if !found {
		return nil, false
	}
	return value.(*View), true
}

// UsersByAction returns actors whose stable room state matches one action.
func (view *View) UsersByAction(action int32) []int64 {
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		if actionMatches(unit, action) {
			result = append(result, unit.EntityKey)
		}
	}
	return result
}

// UsersByName returns players matching an optional exact username.
func (view *View) UsersByName(username string) []int64 {
	result := make([]int64, 0)
	for _, presence := range view.active.Presences() {
		if username == "" || strings.EqualFold(presence.Occupant.Username, username) {
			result = append(result, presence.Unit.EntityKey)
		}
	}
	return result
}

// UsersOnFurniture returns actors occupying any selected furniture footprint.
func (view *View) UsersOnFurniture(targets []record.Target) []int64 {
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		for _, target := range targets {
			item, found := view.active.FurnitureItem(target.ItemID)
			if found && footprintContains(item, unit.Position.Point) {
				result = append(result, unit.EntityKey)
				break
			}
		}
	}
	return result
}

// UsersInGroup returns players belonging to the room's social group.
func (view *View) UsersInGroup() []int64 {
	if view.groups == nil {
		return nil
	}
	result := make([]int64, 0)
	for _, presence := range view.active.Presences() {
		member, loaded := view.groups.IsRoomMember(view.active.ID(), presence.Occupant.PlayerID)
		if loaded && member {
			result = append(result, presence.Unit.EntityKey)
		}
	}
	return result
}

// UsersByHanditem returns actors carrying one hand item.
func (view *View) UsersByHanditem(itemID int32) []int64 {
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		if unit.HandItem == itemID {
			result = append(result, unit.EntityKey)
		}
	}
	return result
}

// UsersByType returns actors matching one protocol unit family.
func (view *View) UsersByType(kind int32) []int64 {
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		if int32(unit.Kind) == kind {
			result = append(result, unit.EntityKey)
		}
	}
	return result
}

// UsersInArea returns actors inside an inclusive room rectangle.
func (view *View) UsersInArea(x1 int, y1 int, x2 int, y2 int) []int64 {
	result := make([]int64, 0)
	for _, unit := range view.active.Units() {
		x, y := int(unit.Position.Point.X), int(unit.Position.Point.Y)
		if x >= x1 && x <= x2 && y >= y1 && y <= y2 {
			result = append(result, unit.EntityKey)
		}
	}
	return result
}

// UsersByTeam returns players assigned to one WIRED game team.
func (view *View) UsersByTeam(team int32) []int64 {
	if view.games == nil {
		return nil
	}
	result := make([]int64, 0)
	for _, presence := range view.active.Presences() {
		actual, found := view.games.Team(view.active.ID(), presence.Occupant.PlayerID)
		if found && actual == team {
			result = append(result, presence.Unit.EntityKey)
		}
	}
	return result
}

// actionMatches compares one stable actor action projection.
func actionMatches(unit roomlive.UnitSnapshot, action int32) bool {
	if action == 4 {
		return !unit.Idle
	}
	if action == 5 {
		return unit.Idle
	}
	for _, status := range unit.Statuses {
		switch action {
		case 6:
			if status.Key == worldunit.StatusSit {
				return true
			}
		case 8:
			if status.Key == worldunit.StatusLay {
				return true
			}
		case 10:
			if status.Key == worldunit.StatusDance {
				value, _ := strconv.ParseInt(status.Value, 10, 32)
				return value > 0
			}
		}
	}
	return action == 7
}
