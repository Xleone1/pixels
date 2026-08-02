package runtime

import "github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"

// Usage stores one allocation-free active room WIRED usage snapshot.
type Usage struct {
	// WiredFurni stores compiled WIRED furniture count.
	WiredFurni int
	// Effects stores compiled effect count.
	Effects int
	// LatestTraceEffects stores effects attempted by the latest completed trace.
	LatestTraceEffects int
	// Stacks stores compiled stack count.
	Stacks int
	// LatestTraceStacks stores stacks visited by the latest completed trace.
	LatestTraceStacks int
	// Delayed stores outstanding delayed work.
	Delayed int
	// ExecutionMicros stores the latest trace duration.
	ExecutionMicros int64
	// MaxEffects stores the per-trace effect budget.
	MaxEffects int
	// MaxStacks stores the per-trace stack budget.
	MaxStacks int
	// MaxDelayed stores the per-room delayed work budget.
	MaxDelayed int
	// MaxVariables stores the per-room durable variable budget.
	MaxVariables int
	// MaxSignals stores the per-trace signal budget.
	MaxSignals int
	// Signals stores signals accepted by the latest completed trace.
	Signals int
	// CompileFailures stores failed generation compilations observed by this process.
	CompileFailures uint64
}

// Context returns the latest completed room execution event for read-only inspection.
func (engine *Engine) Context(roomID int64) (trigger.Event, bool) {
	value, found := engine.rooms.Load(roomID)
	if !found || !engine.config.Enabled {
		return trigger.Event{}, false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	return loaded.lastEvent, loaded.hasLastEvent
}

// Usage returns one active room snapshot without exposing mutable generations.
func (engine *Engine) Usage(roomID int64) (Usage, bool) {
	value, found := engine.rooms.Load(roomID)
	if !found || !engine.config.Enabled {
		return Usage{}, false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	result := Usage{
		WiredFurni: len(loaded.generation.Nodes), Stacks: len(loaded.generation.Stacks),
		Delayed: loaded.delayed, MaxEffects: engine.config.MaxEffectsPerTrace,
		MaxStacks: engine.config.MaxStacksPerTrace, MaxDelayed: engine.config.MaxDelayedPerRoom,
		MaxVariables: engine.config.MaxVariablesPerRoom, MaxSignals: engine.config.MaxSignalsPerTick,
		CompileFailures: engine.metrics.compileFailures.Load(),
	}
	for _, stack := range loaded.generation.Stacks {
		result.Effects += len(stack.Effects)
	}
	if loaded.traceCount > 0 {
		latest := (loaded.traceNext - 1 + len(loaded.traces)) % len(loaded.traces)
		result.ExecutionMicros = loaded.traces[latest].Duration.Microseconds()
		result.Signals = loaded.traces[latest].Signals
		result.LatestTraceEffects = loaded.traces[latest].Effects
		result.LatestTraceStacks = loaded.traces[latest].Stacks
	}
	return result, true
}

// Matches reports whether an active generation has a matching trigger candidate.
func (engine *Engine) Matches(event trigger.Event) bool {
	value, found := engine.rooms.Load(event.RoomID)
	if !found || !engine.config.Enabled {
		return false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	for _, candidate := range loaded.byKind[event.Kind] {
		if engine.matcher.Match(candidate, event) {
			return true
		}
	}
	return false
}

// Conflicts returns trigger sprites incompatible with one effect's actor requirements.
func (engine *Engine) Conflicts(roomID int64, itemID int64) []int32 {
	value, found := engine.rooms.Load(roomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	node := loaded.generation.Nodes[itemID]
	if node == nil {
		return nil
	}
	stack := loaded.generation.Stacks[node.Point]
	result := make([]int32, 0)
	for _, candidate := range stack.Triggers {
		if node.Descriptor.Actor != 0 && candidate.Descriptor.Actor == 0 {
			result = append(result, candidate.SpriteID)
		}
	}
	return result
}
