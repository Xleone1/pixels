package runtime

import (
	"context"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// RoomScheduler queues delayed effects on the existing active-room task queue.
type RoomScheduler struct {
	// rooms resolves active room lifecycle owners.
	rooms *roomlive.Registry
}

// NewRoomScheduler creates a scheduler backed by active rooms.
func NewRoomScheduler(rooms *roomlive.Registry) *RoomScheduler { return &RoomScheduler{rooms: rooms} }

// Schedule queues work only when the room remains active.
func (scheduler *RoomScheduler) Schedule(roomID int64, _ uint64, delay time.Duration, run func(time.Time)) bool {
	if scheduler == nil || scheduler.rooms == nil || run == nil {
		return false
	}
	room, found := scheduler.rooms.Find(roomID)
	if !found {
		return false
	}
	room.Schedule(delay, run)
	return true
}

// releaseDelayed removes one discarded generation's outstanding task gauge.
func (engine *Engine) releaseDelayed(loaded *state) {
	loaded.mutex.Lock()
	delayed := loaded.delayed
	loaded.delayed = 0
	loaded.mutex.Unlock()
	if delayed > 0 {
		engine.metrics.delayedTasks.Add(-int64(delayed))
	}
}

// executeDelayed executes one effect only against the generation that scheduled it.
func (engine *Engine) executeDelayed(ctx context.Context, event trigger.Event, node *configuration.Node, generationID uint64, now time.Time, depth int) error {
	value, found := engine.rooms.Load(event.RoomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	if loaded.generation.ID != generationID {
		return nil
	}
	if loaded.delayed > 0 {
		loaded.delayed--
		engine.metrics.delayedTasks.Add(-1)
	}
	started := time.Now()
	execution := execution{
		context: ctx, state: loaded, event: event, now: now,
		trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started},
	}
	result, err := engine.effects.Execute(ctx, node, event)
	if err == nil {
		engine.recordEffect(result.Status)
		if result.Status == effect.Applied {
			engine.activateNode(ctx, event.RoomID, node)
		}
		execution.trace.Effects++
		engine.applyResult(&execution, result, depth)
		err = engine.run(&execution)
	}
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return err
}
