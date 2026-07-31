package runtime

import (
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
)

// conditionsPass applies whole-stack AND or OR semantics.
func (engine *Engine) conditionsPass(
	execution *execution,
	stack *configuration.Stack,
	resolved selection.Selection,
) (bool, error) {
	if len(stack.Conditions) == 0 {
		return true, nil
	}
	if engine.views == nil {
		return false, nil
	}
	view, found := engine.views.View(execution.event.RoomID)
	if !found {
		return false, nil
	}
	matched := false
	for _, node := range stack.Conditions {
		selectedNode := node
		if resolved.FurniResolved {
			copied := *node
			copied.Targets = resolved.Furni
			selectedNode = &copied
		}
		result, err := engine.conditions.Evaluate(
			selectedNode,
			condition.Context{
				CallbackContext: execution.context,
				Event:           execution.event,
				Now:             execution.now,
				ResetAt:         execution.state.resetAt,
				Effects:         stack.Effects,
				Selection:       resolved,
				Variables:       engine.variables,
			},
			view,
		)
		if err != nil {
			return false, err
		}
		matched = matched || result.Pass
		if stack.Or && result.Pass {
			return true, nil
		}
		if !stack.Or && !result.Pass {
			return false, nil
		}
	}
	return matched, nil
}
