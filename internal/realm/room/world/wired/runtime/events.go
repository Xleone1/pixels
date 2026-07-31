package runtime

import (
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// eventKind maps canonical trigger keys to the runtime index.
func eventKind(key string) trigger.Kind {
	switch key {
	case "wf_trg_enter_room":
		return trigger.EnterRoom
	case "wf_trg_says_something":
		return trigger.Say
	case "wf_trg_walks_on_furni":
		return trigger.WalkOn
	case "wf_trg_walks_off_furni":
		return trigger.WalkOff
	case "wf_trg_state_changed":
		return trigger.StateChanged
	case "wf_trg_collision":
		return trigger.Collision
	case "wf_trg_game_starts":
		return trigger.GameStarted
	case "wf_trg_game_ends":
		return trigger.GameEnded
	case "wf_trg_score_achieved":
		return trigger.ScoreAchieved
	case "wf_trg_bot_reached_stf":
		return trigger.BotReachedFurniture
	case "wf_trg_bot_reached_avtr":
		return trigger.BotReachedAvatar
	case "wf_trg_game_team_win":
		return trigger.TeamWon
	case "wf_trg_game_team_lose":
		return trigger.TeamLost
	case "wf_trg_recv_signal":
		return trigger.ReceiveSignal
	case "wf_trg_leave_room":
		return trigger.LeaveRoom
	case "wf_trg_user_performs_action":
		return trigger.UserPerformsAction
	case "wf_trg_clock_counter":
		return trigger.ClockCounter
	case "wf_trg_var_changed":
		return trigger.VariableChanged
	default:
		kind, _ := timerKind(key)
		return kind
	}
}

// applyResult applies engine-owned effect directives.
func (engine *Engine) applyResult(execution *execution, result effect.Result, depth int) {
	if result.ResetTimers {
		execution.state.resetAt = execution.now
		execution.state.timers = buildTimers(execution.state.generation, execution.now)
	}
	for _, targetID := range result.CallTargets {
		target := execution.state.generation.Nodes[targetID]
		if target != nil {
			execution.enqueue(eventQueue{
				stack: execution.state.generation.Stacks[target.Point],
				event: execution.storeEvent(execution.event), depth: depth + 1,
			})
		}
	}
	for _, derived := range result.Derived {
		if derived.Kind == trigger.ReceiveSignal {
			if execution.signals >= engine.config.MaxSignalsPerTick {
				execution.trace.BudgetExhausted = true
				continue
			}
			execution.signals++
		}
		if derived.RoomID != execution.event.RoomID ||
			execution.queueSize() >= engine.config.MaxEventsPerTrace {
			continue
		}
		for _, candidate := range execution.state.byKind[derived.Kind] {
			if engine.matcher.Match(candidate, derived) {
				execution.enqueue(eventQueue{
					stack:   execution.state.generation.Stacks[candidate.Point],
					trigger: candidate,
					event:   execution.storeEvent(derived), depth: depth + 1,
				})
			}
		}
	}
}

// appendTrace appends one trace to the fixed ring.
func (loaded *state) appendTrace(trace Trace) {
	loaded.traces[loaded.traceNext] = trace
	loaded.traceNext = (loaded.traceNext + 1) % len(loaded.traces)
	if loaded.traceCount < len(loaded.traces) {
		loaded.traceCount++
	}
}

// markVisited reports an existing stack or records it in a small-buffer index.
func (execution *execution) markVisited(point configuration.Point) bool {
	if execution.visitedOverflow != nil {
		if _, found := execution.visitedOverflow[point]; found {
			return true
		}
		execution.visitedOverflow[point] = struct{}{}
		return false
	}
	for _, current := range execution.visited[:execution.visitedCount] {
		if current == point {
			return true
		}
	}
	if execution.visitedCount < len(execution.visited) {
		execution.visited[execution.visitedCount] = point
		execution.visitedCount++
		return false
	}
	execution.visitedOverflow = make(
		map[configuration.Point]struct{}, execution.visitedCount+1,
	)
	for _, current := range execution.visited[:execution.visitedCount] {
		execution.visitedOverflow[current] = struct{}{}
	}
	execution.visitedOverflow[point] = struct{}{}
	return false
}

// enqueue appends one breadth-first request to the inline queue or overflow.
func (execution *execution) enqueue(request eventQueue) {
	if execution.queueLength < len(execution.queue) {
		index := (execution.queueHead + execution.queueLength) % len(execution.queue)
		execution.queue[index] = request
		execution.queueLength++
		return
	}
	execution.queueOverflow = append(execution.queueOverflow, request)
}

// dequeue removes the next breadth-first request.
func (execution *execution) dequeue() (eventQueue, bool) {
	if execution.queueLength == 0 {
		return eventQueue{}, false
	}
	request := execution.queue[execution.queueHead]
	execution.queue[execution.queueHead] = eventQueue{}
	execution.queueHead = (execution.queueHead + 1) % len(execution.queue)
	execution.queueLength--
	if len(execution.queueOverflow) > 0 {
		execution.enqueue(execution.queueOverflow[0])
		execution.queueOverflow = execution.queueOverflow[1:]
	}
	return request, true
}

// queueSize returns all pending inline and overflow requests.
func (execution *execution) queueSize() int {
	return execution.queueLength + len(execution.queueOverflow)
}

// storeEvent retains one derived context in the common inline buffer.
func (execution *execution) storeEvent(event trigger.Event) *trigger.Event {
	if execution.derivedEvents == nil {
		execution.derivedEvents = new([8]trigger.Event)
	}
	if execution.derivedCount < len(*execution.derivedEvents) {
		index := execution.derivedCount
		execution.derivedCount++
		execution.derivedEvents[index] = event
		return &execution.derivedEvents[index]
	}
	stored := &trigger.Event{}
	*stored = event
	execution.derivedOverflow = append(execution.derivedOverflow, stored)
	return stored
}
