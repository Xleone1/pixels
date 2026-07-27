package live

import (
	"context"
	"time"
)

// startLoop starts the room owner goroutine.
func (room *Room) startLoop(ctx context.Context, interval time.Duration, movementPublisher MovementPublisher, doorbellPublisher DoorbellPublisher, cyclePublisher CyclePublisher, doorbellTimeout time.Duration) {
	if movementPublisher == nil && doorbellPublisher == nil && cyclePublisher == nil {
		return
	}
	room.loop.Start(ctx, interval, func(loopCtx context.Context) {
		started := time.Now()
		failed := false
		movements := room.Tick()
		if len(movements) > 0 && movementPublisher != nil {
			failed = movementPublisher(loopCtx, room, movements) != nil || failed
		}
		expired := room.SweepDoorbell(time.Now(), doorbellTimeout)
		if len(expired) > 0 && doorbellPublisher != nil {
			failed = doorbellPublisher(loopCtx, room, expired) != nil || failed
		}
		if cyclePublisher != nil {
			failed = cyclePublisher(loopCtx, room, time.Now()) != nil || failed
		}
		room.recordTick(started, failed)
	})
}

// stopLoop stops the room owner goroutine.
func (room *Room) stopLoop() { room.loop.Stop() }
