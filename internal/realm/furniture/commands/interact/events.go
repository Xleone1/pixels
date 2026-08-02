package interact

import (
	"context"

	furnitureclicked "github.com/niflaot/pixels/internal/realm/furniture/events/clicked"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	"github.com/niflaot/pixels/pkg/bus"
)

// publishClick emits one validated interaction attempt before specialized behavior.
func (handler Handler) publishClick(ctx context.Context, playerID int64, itemID int64, roomID int64) error {
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(ctx, bus.Event{Name: furnitureclicked.Name, Payload: furnitureclicked.Payload{
		PlayerID: playerID, ItemID: itemID, RoomID: roomID,
	}})
}

// publish emits one successful generic furniture use event.
func (handler Handler) publish(ctx context.Context, playerID int64, itemID int64, roomID int64) error {
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(ctx, bus.Event{Name: furnitureused.Name, Payload: furnitureused.Payload{
		PlayerID: playerID, ItemID: itemID, RoomID: roomID,
	}})
}
