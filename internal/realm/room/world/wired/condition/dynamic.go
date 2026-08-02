package condition

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// altitudeMatches checks every configured target against one inclusive range.
func altitudeMatches(node *configuration.Node, view View) (Result, error) {
	if len(node.Targets) == 0 {
		return Result{}, nil
	}
	minimum, maximum := first(node.Parameters.Values), value(node.Parameters.Values, 1)
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	for _, target := range node.Targets {
		height, found := view.FurnitureAltitude(target.ItemID)
		if !found || height < minimum || height > maximum {
			return Result{Pass: false, Valid: found}, nil
		}
	}
	return Result{Pass: true, Valid: true}, nil
}

// variableExists checks one resolved variable key.
func variableExists(node *configuration.Node, context Context) (Result, error) {
	current, found, err := resolveVariable(node, context)
	if err != nil || !found {
		return Result{Valid: err == nil}, err
	}
	if len(node.Parameters.Values) >= 3 {
		return Result{
			Pass:  current.IntValue == int64(node.Parameters.Values[2]),
			Valid: true,
		}, nil
	}
	return Result{Pass: true, Valid: true}, nil
}

// variableValueMatches compares a value or age through one stable operator.
func variableValueMatches(
	node *configuration.Node,
	context Context,
	age bool,
) (Result, error) {
	current, found, err := resolveVariable(node, context)
	if err != nil || !found {
		return Result{Valid: err == nil}, err
	}
	actual := current.IntValue
	if age {
		actual = int64(context.Now.Sub(current.CreatedAt) / time.Second)
	}
	expected := int64(value(node.Parameters.Values, 2))
	return Result{
		Pass:  compare(actual, expected, value(node.Parameters.Values, 1)),
		Valid: true,
	}, nil
}

// resolveVariable resolves scope and target from the current event and selection.
func resolveVariable(
	node *configuration.Node,
	context Context,
) (variable.Value, bool, error) {
	if node.Parameters.Text == "" {
		return variable.Value{}, false, nil
	}
	scope := variable.Scope(first(node.Parameters.Values))
	if scope == variable.ScopeContext {
		current, found := variable.ResolveContext(context.Event, node.Parameters.Text, context.Now)
		return current, found, nil
	}
	if context.Variables == nil {
		return variable.Value{}, false, nil
	}
	scopeID := context.Event.RoomID
	switch scope {
	case variable.ScopeFurni:
		if len(context.Selection.Furni) == 0 {
			return variable.Value{}, false, nil
		}
		scopeID = context.Selection.Furni[0].ItemID
	case variable.ScopeUser:
		scopeID = context.Event.PlayerID
		if scopeID == 0 {
			scopeID = context.Event.ActorID
		}
	case variable.ScopeRoom:
	case variable.ScopeReference:
		if context.Event.ReferenceRoomID <= 0 ||
			context.Event.ReferenceVariable == "" {
			return variable.Value{}, false, nil
		}
		roomID := context.Event.ReferenceRoomID
		return context.Variables.GetDurable(
			context.CallbackContext,
			roomID,
			variable.ScopeRoom,
			roomID,
			context.Event.ReferenceVariable,
		)
	default:
		return variable.Value{}, false, nil
	}
	current, found := context.Variables.Get(
		context.Event.RoomID, scope, scopeID, node.Parameters.Text,
	)
	return current, found, nil
}

// compare applies equal, greater-than, less-than, greater-or-equal, or less-or-equal.
func compare(actual int64, expected int64, operation int32) bool {
	switch operation {
	case 1:
		return actual > expected
	case 2:
		return actual < expected
	case 3:
		return actual >= expected
	case 4:
		return actual <= expected
	case 5:
		return actual != expected
	default:
		return actual == expected
	}
}

// value returns one integer setting or zero.
func value(values []int32, index int) int32 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}
