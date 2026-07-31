package selection

import (
	"slices"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// metric binds one selection index to its sortable variable value.
type metric struct {
	// index stores the original stable selection index.
	index int
	// value stores the warmed variable assignment.
	value variable.Value
}

// FilterVariables applies the official variable ordering extras to one selection.
func FilterVariables(
	stack *configuration.Stack,
	selected Selection,
	event trigger.Event,
	variables *variable.Service,
) Selection {
	if stack == nil || variables == nil || len(stack.Extras) == 0 {
		return selected
	}
	for _, extra := range stack.Extras {
		switch extra.Descriptor.Key {
		case "wf_xtra_filter_users_by_var":
			selected.Actors = filterActors(extra, selected.Actors, event, variables)
		case "wf_xtra_filter_furni_by_var":
			selected.Furni = filterFurni(extra, selected.Furni, event, variables)
		}
	}
	return selected
}

// filterActors retains actors with a variable and orders the bounded result.
func filterActors(
	extra *configuration.Node,
	actors []int64,
	event trigger.Event,
	variables *variable.Service,
) []int64 {
	metrics := make([]metric, 0, len(actors))
	for index, actorID := range actors {
		value, found := variables.Get(
			event.RoomID, variable.ScopeUser, actorID, extra.Parameters.Text,
		)
		if found {
			metrics = append(metrics, metric{index: index, value: value})
		}
	}
	sortMetrics(metrics, parameterValue(extra, 0))
	amount := filterAmount(extra, event, variables)
	if amount > len(metrics) {
		amount = len(metrics)
	}
	result := actors[:0]
	for _, current := range metrics[:amount] {
		result = append(result, actors[current.index])
	}
	return result
}

// filterFurni retains furniture with a variable and orders the bounded result.
func filterFurni(
	extra *configuration.Node,
	furni []record.Target,
	event trigger.Event,
	variables *variable.Service,
) []record.Target {
	metrics := make([]metric, 0, len(furni))
	for index, target := range furni {
		value, found := variables.Get(
			event.RoomID, variable.ScopeFurni, target.ItemID, extra.Parameters.Text,
		)
		if found {
			metrics = append(metrics, metric{index: index, value: value})
		}
	}
	sortMetrics(metrics, parameterValue(extra, 0))
	amount := filterAmount(extra, event, variables)
	if amount > len(metrics) {
		amount = len(metrics)
	}
	result := furni[:0]
	for _, current := range metrics[:amount] {
		result = append(result, furni[current.index])
	}
	return result
}

// sortMetrics applies the six Polaris variable sort modes with stable ties.
func sortMetrics(metrics []metric, mode int32) {
	slices.SortStableFunc(metrics, func(left metric, right metric) int {
		comparison := compareMetric(left.value, right.value, mode)
		if comparison != 0 {
			return comparison
		}
		return left.index - right.index
	})
}

// compareMetric compares numeric value, creation time, or update time.
func compareMetric(left variable.Value, right variable.Value, mode int32) int {
	switch mode {
	case 1:
		return compareInt(left.IntValue, right.IntValue)
	case 2:
		return left.CreatedAt.Compare(right.CreatedAt)
	case 3:
		return right.CreatedAt.Compare(left.CreatedAt)
	case 4:
		return left.UpdatedAt.Compare(right.UpdatedAt)
	case 5:
		return right.UpdatedAt.Compare(left.UpdatedAt)
	default:
		return compareInt(right.IntValue, left.IntValue)
	}
}

// compareInt returns the ordering of two signed integers without overflow.
func compareInt(left int64, right int64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

// filterAmount resolves constant or variable-driven result limits.
func filterAmount(
	extra *configuration.Node,
	event trigger.Event,
	variables *variable.Service,
) int {
	if parameterValue(extra, 1) != 1 {
		amount := int32(1)
		if len(extra.Parameters.Values) > 2 {
			amount = parameterValue(extra, 2)
		}
		return boundedAmount(int(amount))
	}
	scope := variable.Scope(parameterValue(extra, 3) + 1)
	scopeID := event.ActorID
	if scope == variable.ScopeRoom || scope == variable.ScopeReference {
		scopeID = event.RoomID
	}
	value, found := variables.Get(
		event.RoomID, scope, scopeID, extra.Parameters.Message,
	)
	if !found {
		return 0
	}
	return boundedAmount(int(value.IntValue))
}

// boundedAmount clamps variable-filter output to the official safety maximum.
func boundedAmount(value int) int {
	if value < 0 {
		return 0
	}
	if value > 10000 {
		return 10000
	}
	return value
}

// parameterValue returns one integer setting or zero.
func parameterValue(node *configuration.Node, index int) int32 {
	if index < 0 || index >= len(node.Parameters.Values) {
		return 0
	}
	return node.Parameters.Values[index]
}
