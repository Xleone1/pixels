package interact

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnitureclicked "github.com/niflaot/pixels/internal/realm/furniture/events/clicked"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestHandlePublishesValidatedClickBeforeUnsupportedBehavior verifies click triggers do not depend on a state mutation.
func TestHandlePublishesValidatedClickBeforeUnsupportedBehavior(t *testing.T) {
	handler, connection, _, _ := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0),
		Definition: worldfurniture.Definition{InteractionType: "unknown", Width: 1, Length: 1},
	})
	clicked := furnitureclicked.Payload{}
	if _, err := handler.Events.(*bus.Bus).Subscribe(furnitureclicked.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		clicked, _ = event.Payload.(furnitureclicked.Payload)
		return nil
	}); err != nil {
		t.Fatalf("subscribe click: %v", err)
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("interact: %v", err)
	}
	if clicked.PlayerID != 7 || clicked.ItemID != 1 || clicked.RoomID != 9 {
		t.Fatalf("unexpected click payload %#v", clicked)
	}
}
