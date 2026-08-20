package walk

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// walkEventsForTest redirects or cancels one movement.
type walkEventsForTest struct {
	// x stores the replacement x coordinate.
	x int
	// y stores the replacement y coordinate.
	y int
	// cancelled stores the veto decision.
	cancelled bool
	// calls stores dispatch count.
	calls int
}

// HasListeners reports a movement observer.
func (*walkEventsForTest) HasListeners(name string) bool { return name == "room.unit.move" }

// DispatchRoomUnitMove returns the configured plugin decision.
func (events *walkEventsForTest) DispatchRoomUnitMove(_ context.Context, _ sdkplayer.Player, _ int64, _ int, _ int) (int, int, bool) {
	events.calls++
	return events.x, events.y, events.cancelled
}

// TestHandleMoveErrorSettlesActiveMovement verifies soft misses do not broadcast a snapping status.
func TestHandleMoveErrorSettlesActiveMovement(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registeredConnectionForWalkTest(t, connections, "conn")
	handler.Connections = connections
	room, _ := handler.Runtime.Find(9)
	if _, err := room.MoveTo(7, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("start movement: %v", err)
	}
	if err := handler.handleMoveError(context.Background(), room, 7, grid.MustPoint(1, 0), worldpath.ErrInvalidGoal); err != nil {
		t.Fatalf("handle movement miss: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no immediate snapping packet, got %#v", *sent)
	}
	movements := room.Tick()
	if len(movements) != 1 || !movements[0].Settled || movements[0].Moved {
		t.Fatalf("expected deferred neutral settlement, got %#v", movements)
	}
}

// TestHandleFacesOccupiedTarget verifies occupied targets do not disconnect clients.
func TestHandleFacesOccupiedTarget(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registeredConnectionForWalkTest(t, connections, "conn")
	handler.Connections = connections
	room, _ := handler.Runtime.Find(9)
	if _, err := room.Join(roomlive.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join other: %v", err)
	}
	if _, err := room.MoveTo(8, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("move other: %v", err)
	}
	if movements := room.Tick(); len(movements) != 1 {
		t.Fatalf("expected other movement %#v", movements)
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), X: 1, Y: 0},
	})
	if err != nil {
		t.Fatalf("handle occupied walk: %v", err)
	}
	units := room.Units()
	if len(units) != 2 || units[0].BodyRotation != worldunit.RotationEast {
		t.Fatalf("expected player facing target %#v", units)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 1640 {
		t.Fatalf("expected status packet, got %#v", *sent)
	}
}

// TestPluginWalkRedirectsAreRevalidatedAndCancellationStopsMovement verifies the movement seam.
func TestPluginWalkRedirectsAreRevalidatedAndCancellationStopsMovement(t *testing.T) {
	t.Run("redirected", func(t *testing.T) {
		handler, player := handlerForTest(t)
		if err := player.EnterRoom(9); err != nil {
			t.Fatal(err)
		}
		events := &walkEventsForTest{x: 2}
		handler.PluginEvents = events
		if err := handler.Handle(context.Background(), commandEnvelopeForWalkTest(1, 0)); err != nil {
			t.Fatal(err)
		}
		room, _ := handler.Runtime.Find(9)
		room.Tick()
		room.Tick()
		unit, _ := room.Unit(7)
		if unit.Position.Point != grid.MustPoint(2, 0) || events.calls != 1 {
			t.Fatalf("unit=%#v calls=%d", unit, events.calls)
		}
	})
	t.Run("redirect revalidated", func(t *testing.T) {
		handler, player := handlerForTest(t)
		if err := player.EnterRoom(9); err != nil {
			t.Fatal(err)
		}
		events := &walkEventsForTest{x: -1}
		handler.PluginEvents = events
		err := handler.Handle(context.Background(), commandEnvelopeForWalkTest(1, 0))
		if !errors.Is(err, ErrInvalidTarget) || events.calls != 1 {
			t.Fatalf("calls=%d err=%v", events.calls, err)
		}
	})
	t.Run("cancelled", func(t *testing.T) {
		handler, player := handlerForTest(t)
		if err := player.EnterRoom(9); err != nil {
			t.Fatal(err)
		}
		events := &walkEventsForTest{x: 1, cancelled: true}
		handler.PluginEvents = events
		if err := handler.Handle(context.Background(), commandEnvelopeForWalkTest(1, 0)); err != nil {
			t.Fatal(err)
		}
		room, _ := handler.Runtime.Find(9)
		unit, _ := room.Unit(7)
		if unit.Moving || events.calls != 1 {
			t.Fatalf("unit=%#v calls=%d", unit, events.calls)
		}
	})
}

// commandEnvelopeForWalkTest creates one walk command envelope.
func commandEnvelopeForWalkTest(x int, y int) command.Envelope[Command] {
	return command.Envelope[Command]{Command: Command{Handler: connectionForTest(), X: x, Y: y}}
}
