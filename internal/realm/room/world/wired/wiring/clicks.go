package wiring

import (
	"context"

	furnitureclicked "github.com/niflaot/pixels/internal/realm/furniture/events/clicked"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	avatarclicked "github.com/niflaot/pixels/internal/realm/room/world/events/avatarclicked"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/pkg/bus"
)

// avatarClickedHandler emits validated user-on-user clicks.
func avatarClickedHandler(rooms *roomlive.Registry, engine *wiredruntime.Engine) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(avatarclicked.Payload)
		if ok {
			scheduleEvent(rooms, engine, trigger.Event{
				Kind: trigger.AvatarClicked, RoomID: payload.RoomID,
				ActorKind: trigger.ActorPlayer, ActorID: payload.PlayerID,
				PlayerID: payload.PlayerID, TargetPlayerID: payload.TargetPlayerID,
			})
		}
		return nil
	}
}

// clickedHandler emits validated user furniture clicks.
func clickedHandler(rooms *roomlive.Registry, engine *wiredruntime.Engine) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnitureclicked.Payload)
		if ok {
			wiredEvent := unitEvent(rooms, trigger.FurnitureClicked, payload.RoomID, payload.PlayerID, payload.ItemID)
			scheduleEvent(rooms, engine, wiredEvent)
			if clickedFloorTile(rooms, payload.RoomID, payload.ItemID) {
				wiredEvent.Kind = trigger.FloorTileClicked
				scheduleEvent(rooms, engine, wiredEvent)
			}
		}
		return nil
	}
}

// clickedFloorTile reports whether one validated click targeted a floor-tile source.
func clickedFloorTile(rooms *roomlive.Registry, roomID int64, itemID int64) bool {
	active, found := rooms.Find(roomID)
	if !found {
		return false
	}
	item, found := active.FurnitureItem(itemID)
	return found && (item.Definition.InteractionType == "room_invisible_click_tile" ||
		item.Definition.InteractionType == "wf_trg_user_clicks_tile")
}
