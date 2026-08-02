package room

import (
	"strconv"
	"strings"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomaction "github.com/niflaot/pixels/internal/realm/room/world/action"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
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
	if !roomaction.ValidWiredAction(action) {
		return false
	}
	if action == roomaction.WiredActionJump {
		return !unit.Idle
	}
	if action == roomaction.WiredActionRespect {
		return unit.Idle
	}
	for _, status := range unit.Statuses {
		switch action {
		case roomaction.WiredActionSit:
			if status.Key == worldunit.StatusSit {
				return true
			}
		case roomaction.WiredActionLay:
			if status.Key == worldunit.StatusLay {
				return true
			}
		case roomaction.WiredActionDance:
			if status.Key == worldunit.StatusDance {
				value, _ := strconv.ParseInt(status.Value, 10, 32)
				return value > 0
			}
		}
	}
	return action == roomaction.WiredActionStand
}

// PerformsAction reports whether the event actor matches one configured action and variant.
func (view *View) PerformsAction(event trigger.Event, values []int32) (bool, bool, error) {
	action := int32(0)
	if len(values) > 0 {
		action = values[0]
	}
	if !roomaction.ValidWiredAction(action) || event.ActorID == 0 {
		return false, false, nil
	}
	if event.Kind == trigger.UserPerformsAction && event.Action == action {
		return actionVariantMatches(action, event.ActionValue, values), true, nil
	}
	unit, found := view.active.UnitMotion(event.ActorID)
	if !found {
		return false, false, nil
	}
	if unit.Kind == worldunit.KindPlayer {
		unit, found = view.active.Unit(unit.PlayerID)
	}
	return found && actionMatches(unit, action) && stableActionVariantMatches(unit, action, values), found, nil
}

// actionVariantMatches compares optional sign and dance filters.
func actionVariantMatches(action int32, actionValue int32, values []int32) bool {
	if action == roomaction.WiredActionSign && valueAt(values, 1) != 0 {
		return actionValue == valueAt(values, 2)
	}
	if action == roomaction.WiredActionDance && valueAt(values, 3) != 0 {
		return actionValue == valueAt(values, 4)
	}
	return true
}

// stableActionVariantMatches compares persistent action variants.
func stableActionVariantMatches(unit roomlive.UnitSnapshot, action int32, values []int32) bool {
	if action != roomaction.WiredActionDance || valueAt(values, 3) == 0 {
		return action != roomaction.WiredActionSign || valueAt(values, 1) == 0
	}
	for _, status := range unit.Statuses {
		if status.Key == worldunit.StatusDance {
			current, _ := strconv.ParseInt(status.Value, 10, 32)
			return int32(current) == valueAt(values, 4)
		}
	}
	return false
}

// valueAt returns one optional action parameter.
func valueAt(values []int32, index int) int32 {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return 0
}
