package effect

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// variableEffect reports whether a descriptor mutates WIRED variables.
func variableEffect(key string) bool {
	switch key {
	case "wf_act_give_var", "wf_act_remove_var", "wf_act_change_var_val":
		return true
	default:
		return false
	}
}

// executeVariable applies one variable mutation and emits change events.
func (executor *Executor) executeVariable(ctx context.Context, node *configuration.Node, event trigger.Event) (Result, error) {
	if executor.services.Variables == nil || node.Parameters.Text == "" {
		return Result{Status: Blocked}, nil
	}
	scope := variable.Scope(parameter(node.Parameters.Values, 0))
	roomID, name := event.RoomID, node.Parameters.Text
	if scope == variable.ScopeReference {
		if event.ReferenceRoomID <= 0 || event.ReferenceVariable == "" {
			return Result{Status: Skipped}, nil
		}
		roomID, name = event.ReferenceRoomID, event.ReferenceVariable
		scope = variable.ScopeRoom
	}
	scopeIDs := variableTargets(scope, node, event, roomID)
	if len(scopeIDs) == 0 {
		return Result{Status: Skipped}, nil
	}
	derived := make([]trigger.Event, 0, len(scopeIDs))
	for _, scopeID := range scopeIDs {
		next, previous, changed, err := executor.mutateVariable(
			ctx, node, event, roomID, scope, scopeID, name,
		)
		if err != nil {
			return Result{Status: Blocked}, err
		}
		if changed {
			derived = append(derived, variableEvent(event, next, previous))
		}
	}
	if len(derived) == 0 {
		return Result{Status: Skipped}, nil
	}
	return Result{Status: Applied, Derived: derived}, nil
}

// mutateVariable applies one descriptor to one exact scope target.
func (executor *Executor) mutateVariable(
	ctx context.Context,
	node *configuration.Node,
	event trigger.Event,
	roomID int64,
	scope variable.Scope,
	scopeID int64,
	name string,
) (variable.Value, int64, bool, error) {
	value := variable.Value{
		RoomID: roomID, Scope: scope, ScopeID: scopeID, Name: name,
	}
	switch node.Descriptor.Key {
	case "wf_act_give_var":
		value.IntValue = int64(parameter(node.Parameters.Values, 1))
		value.StringValue = node.Parameters.Message
		previous, found, findErr := executor.services.Variables.GetDurable(
			ctx, roomID, scope, scopeID, name,
		)
		if findErr != nil {
			return variable.Value{}, 0, false, findErr
		}
		saved, err := executor.services.Variables.Set(ctx, value)
		return saved, previous.IntValue, !found || previous.IntValue != saved.IntValue || previous.StringValue != saved.StringValue, err
	case "wf_act_change_var_val":
		saved, previous, err := executor.services.Variables.Change(ctx, value, int64(parameter(node.Parameters.Values, 1)))
		return saved, previous, err == nil && previous != saved.IntValue, err
	case "wf_act_remove_var":
		previous, deleted, err := executor.services.Variables.Delete(
			ctx, roomID, scope, scopeID, name,
		)
		return previous, previous.IntValue, deleted, err
	default:
		return variable.Value{}, 0, false, nil
	}
}

// variableTargets resolves concrete target ids without room-state reads.
func variableTargets(
	scope variable.Scope,
	node *configuration.Node,
	event trigger.Event,
	roomID int64,
) []int64 {
	switch scope {
	case variable.ScopeFurni:
		values := make([]int64, 0, len(node.Targets))
		for _, target := range node.Targets {
			values = append(values, target.ItemID)
		}
		return values
	case variable.ScopeUser:
		target := event.PlayerID
		if target == 0 {
			target = event.ActorID
		}
		if target > 0 {
			return []int64{target}
		}
	case variable.ScopeRoom:
		return []int64{roomID}
	}
	return nil
}

// variableEvent creates one derived variable-change trigger event.
func variableEvent(event trigger.Event, value variable.Value, previous int64) trigger.Event {
	event.ID, event.Kind = 0, trigger.VariableChanged
	event.VariableName, event.VariableValue = value.Name, value.IntValue
	event.PreviousVariableValue = previous
	return event
}

// parameter returns one integer setting or zero.
func parameter(values []int32, index int) int32 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}
