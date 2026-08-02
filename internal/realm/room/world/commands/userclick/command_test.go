package userclick

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	avatarclicked "github.com/niflaot/pixels/internal/realm/room/world/events/avatarclicked"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestHandlePublishesValidatedPlayerClicks verifies room-local unit resolution.
func TestHandlePublishesValidatedPlayerClicks(t *testing.T) {
	handler, actor, active, local := userClickFixture(t)
	var captured avatarclicked.Payload
	_, err := local.Subscribe(avatarclicked.Name, 0, func(_ context.Context, event bus.Event) error {
		captured, _ = event.Payload.(avatarclicked.Payload)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	target, found := active.Unit(8)
	if !found {
		t.Fatal("target unit")
	}
	envelope := command.Envelope[Command]{Command: Command{Handler: userClickConnection(), RoomUnitID: target.UnitID}}
	if err = handler.Handle(context.Background(), envelope); err != nil {
		t.Fatal(err)
	}
	if captured.RoomID != active.ID() || captured.PlayerID != actor.ID() || captured.TargetPlayerID != 8 {
		t.Fatalf("captured=%+v", captured)
	}
	captured = avatarclicked.Payload{}
	envelope.Command.RoomUnitID = 999
	if err = handler.Handle(context.Background(), envelope); err != nil || captured != (avatarclicked.Payload{}) {
		t.Fatalf("invalid target captured=%+v err=%v", captured, err)
	}
	actorUnit, found := active.Unit(actor.ID())
	if !found {
		t.Fatal("actor unit")
	}
	envelope.Command.RoomUnitID = actorUnit.UnitID
	if err = handler.Handle(context.Background(), envelope); err != nil || captured != (avatarclicked.Payload{}) {
		t.Fatalf("self target captured=%+v err=%v", captured, err)
	}
}

// TestHandleRequiresCurrentActiveRoom verifies stale session state fails closed.
func TestHandleRequiresCurrentActiveRoom(t *testing.T) {
	handler, actor, _, _ := userClickFixture(t)
	actor.LeaveRoom()
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: userClickConnection(), RoomUnitID: 1}})
	if !errors.Is(err, ErrPlayerNotInRoom) {
		t.Fatalf("missing room error=%v", err)
	}
	if err = actor.EnterRoom(999); err != nil {
		t.Fatal(err)
	}
	err = handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: userClickConnection(), RoomUnitID: 1}})
	if !errors.Is(err, roomlive.ErrRoomNotFound) {
		t.Fatalf("inactive room error=%v", err)
	}
	if (Command{}).CommandName() != Name {
		t.Fatal("unstable command name")
	}
}

// userClickFixture creates two authenticated player units in one active room.
func userClickFixture(t *testing.T) (Handler, *playerlive.Player, *roomlive.Room, *bus.Bus) {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	actor := addClickPlayer(t, players, bindings, 7, "actor", "conn")
	_ = addClickPlayer(t, players, bindings, 8, "target", "target")
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	for _, occupant := range []roomlive.Occupant{
		{PlayerID: 7, Username: "actor", ConnectionID: "conn", ConnectionKind: "websocket"},
		{PlayerID: 8, Username: "target", ConnectionID: "target", ConnectionKind: "websocket"},
	} {
		if _, err = rooms.Join(context.Background(), active.ID(), occupant); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = active.AddEntity(-1, 1, worldunit.KindBot, worldpath.Position{Point: grid.MustPoint(1, 0)}, worldunit.RotationSouth); err != nil {
		t.Fatal(err)
	}
	if err = actor.EnterRoom(active.ID()); err != nil {
		t.Fatal(err)
	}
	local := bus.New()
	t.Cleanup(func() {
		_ = local.Close()
		_, _, _ = rooms.Close(context.Background(), active.ID())
	})
	return Handler{Players: players, Bindings: bindings, Runtime: rooms, Events: local}, actor, active, local
}

// addClickPlayer adds one bound live player to a fixture.
func addClickPlayer(t *testing.T, players *playerlive.Registry, bindings *binding.Registry, id int64, username string, connectionID string) *playerlive.Player {
	t.Helper()
	peer, err := playerlive.NewSessionPeer(netconn.ID(connectionID), "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: username}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	if err = bindings.Add(binding.Binding{PlayerID: id, ConnectionID: netconn.ID(connectionID), ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	return player
}

// userClickConnection creates the actor packet context.
func userClickConnection() netconn.Context {
	return netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}
}
