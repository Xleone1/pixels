package runtime

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// processLocked executes every matching candidate with the room state locked.
func (engine *Engine) processLocked(ctx context.Context, loaded *state, event trigger.Event, now time.Time) (Trace, error) {
	candidates := loaded.byKind[event.Kind]
	if len(candidates) == 0 {
		return Trace{}, nil
	}
	started := time.Now()
	execution := execution{
		context: ctx, state: loaded, event: event, now: now,
		trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started},
	}
	for _, node := range candidates {
		if engine.matcher.Match(node, event) && engine.selectorSourceMatches(ctx, loaded, node, event) {
			execution.enqueue(eventQueue{
				stack: loaded.generation.Stacks[node.Point], trigger: node,
			})
		}
	}
	err := engine.run(&execution)
	execution.trace.Signals = execution.signals
	loaded.lastEvent, loaded.hasLastEvent = execution.event, true
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return execution.trace, err
}

// selectorSourceMatches resolves click-trigger selector output only for source mode 200.
func (engine *Engine) selectorSourceMatches(ctx context.Context, loaded *state, node *configuration.Node, event trigger.Event) bool {
	if !usesClickSelectorSource(node) {
		return true
	}
	stack := loaded.generation.Stacks[node.Point]
	probe := execution{context: ctx, state: loaded, event: event}
	resolved, err := engine.resolveSelection(&probe, stack)
	if err != nil || !resolved.FurniResolved {
		return false
	}
	for _, target := range resolved.Furni {
		if target.ItemID == event.SourceItem {
			return true
		}
	}
	return false
}

// processTriggerLocked executes one already-selected timer trigger.
func (engine *Engine) processTriggerLocked(ctx context.Context, loaded *state, node *configuration.Node, event trigger.Event, now time.Time) (Trace, error) {
	started := time.Now()
	execution := execution{
		context: ctx, state: loaded, event: event, now: now,
		trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started},
	}
	execution.enqueue(eventQueue{
		stack: loaded.generation.Stacks[node.Point], trigger: node,
	})
	err := engine.run(&execution)
	execution.trace.Signals = execution.signals
	loaded.lastEvent, loaded.hasLastEvent = execution.event, true
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return execution.trace, err
}

// run drains one breadth-first trace within configured budgets.
func (engine *Engine) run(execution *execution) error {
	var result error
	for execution.queueSize() > 0 {
		request, _ := execution.dequeue()
		if request.stack == nil {
			continue
		}
		execution.trace.StackPoint = request.stack.Point
		if request.depth > engine.config.MaxCallDepth {
			execution.trace.BudgetExhausted = true
			execution.trace.BudgetCode = "CALL_DEPTH"
			continue
		}
		if request.event == nil {
			execution.event.ReferenceRoomID = 0
			execution.event.ReferenceVariable = ""
		} else {
			execution.event = *request.event
		}
		execution.event = withVariableReference(request.stack, execution.event)
		if execution.markVisited(request.stack.Point) {
			continue
		}
		if execution.trace.Stacks >= engine.config.MaxStacksPerTrace {
			execution.trace.BudgetExhausted = true
			execution.trace.BudgetCode = "STACK_CAP"
			break
		}
		execution.trace.Stacks++
		engine.activateNode(execution.context, execution.event.RoomID, request.trigger)
		resolved, err := engine.resolveSelection(execution, request.stack)
		result = joined(result, err)
		if err != nil {
			engine.metrics.stackResults[2].Add(1)
			continue
		}
		resolved = selection.FilterVariables(
			request.stack, resolved, execution.event, engine.variables,
		)
		passed, err := engine.conditionsPass(execution, request.stack, resolved)
		result = joined(result, err)
		if err != nil {
			engine.metrics.stackResults[2].Add(1)
		} else if passed {
			engine.metrics.stackResults[0].Add(1)
		} else {
			engine.metrics.stackResults[1].Add(1)
		}
		if err != nil || !passed {
			continue
		}
		engine.activateNodes(execution.context, execution.event.RoomID, request.stack.Extras)
		if engine.extras != nil {
			result = joined(result, engine.extras.ExecuteExtras(
				execution.context, request.stack.Extras, execution.event, resolved, execution.now,
			))
		}
		effects := engine.selectEffects(execution.state, request.stack, execution.event.ID)
		for _, node := range effects {
			if !engine.executeSelected(execution, node, request.depth, resolved, &result) {
				break
			}
		}
	}
	return result
}

// resolveSelection evaluates selectors only when the stack owns any.
func (engine *Engine) resolveSelection(execution *execution, stack *configuration.Stack) (selection.Selection, error) {
	if len(stack.Selectors) == 0 || engine.selections == nil {
		return selection.Selection{}, nil
	}
	provider, supported := engine.views.(selection.Provider)
	if !supported {
		return selection.Selection{}, nil
	}
	view, found := provider.SelectionView(execution.event.RoomID)
	if !found {
		return selection.Selection{}, nil
	}
	return engine.selections.ResolveStack(execution.context, stack, execution.event, execution.state.generation, view)
}

// executeSelected executes one effect for each selected actor with dynamic furniture targets.
func (engine *Engine) executeSelected(execution *execution, node *configuration.Node, depth int, resolved selection.Selection, aggregate *error) bool {
	selectedNode := node
	if resolved.FurniResolved {
		copied := *node
		copied.Targets = resolved.Furni
		selectedNode = &copied
	}
	event := execution.event
	event.ActorIDs, event.FurniTargets = resolved.Actors, resolved.Furni
	if resolved.ActorsResolved && len(resolved.Actors) == 0 &&
		node.Descriptor.Actor != registry.ActorOptional {
		return true
	}
	if len(resolved.Actors) == 0 {
		return engine.executeOne(execution, selectedNode, event, depth, aggregate)
	}
	for _, actorID := range resolved.Actors {
		event.ActorID, event.PlayerID = actorID, actorID
		if !engine.executeOne(execution, selectedNode, event, depth, aggregate) {
			return false
		}
	}
	return true
}

// executeOne applies the trace budget around one resolved effect invocation.
func (engine *Engine) executeOne(execution *execution, node *configuration.Node, event trigger.Event, depth int, aggregate *error) bool {
	if execution.trace.Effects >= engine.config.MaxEffectsPerTrace {
		execution.trace.BudgetExhausted = true
		execution.trace.BudgetCode = "EFFECT_CAP"
		return false
	}
	execution.trace.Effects++
	original := execution.event
	execution.event = event
	*aggregate = joined(*aggregate, engine.execute(execution, node, depth))
	execution.event = original
	return true
}

// selectEffects applies random or unseen stack selectors.
func (engine *Engine) selectEffects(loaded *state, stack *configuration.Stack, eventID uint64) []*configuration.Node {
	if len(stack.Effects) <= 1 {
		return stack.Effects
	}
	if stack.Unseen {
		index := loaded.unseen[stack.Point] % len(stack.Effects)
		loaded.unseen[stack.Point] = index + 1
		return stack.Effects[index : index+1]
	}
	if stack.Random {
		index := randomIndex(eventID, stack.Point, len(stack.Effects))
		return stack.Effects[index : index+1]
	}
	return stack.Effects
}

// execute runs or schedules one effect and enqueues its directives.
func (engine *Engine) execute(execution *execution, node *configuration.Node, depth int) error {
	if node.Delay > 0 {
		if engine.scheduler == nil || execution.state.delayed >= engine.config.MaxDelayedPerRoom {
			return nil
		}
		generationID := execution.state.generation.ID
		event := execution.event
		execution.state.delayed++
		engine.metrics.delayedTasks.Add(1)
		if !engine.scheduler.Schedule(event.RoomID, generationID, node.Delay, func(now time.Time) {
			_ = engine.executeDelayed(context.Background(), event, node, generationID, now, depth)
		}) {
			execution.state.delayed--
			engine.metrics.delayedTasks.Add(-1)
		}
		return nil
	}
	result, err := engine.effects.Execute(execution.context, node, execution.event)
	if err != nil {
		return err
	}
	engine.recordEffect(result.Status)
	if result.Status == effect.Applied {
		engine.activateNode(execution.context, execution.event.RoomID, node)
	}
	engine.applyResult(execution, result, depth)
	return nil
}
