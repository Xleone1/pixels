package event

import (
	"context"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	"go.uber.org/zap"
)

// TestRoomDispatchersApplyMutationsAndCancellation verifies every room gate.
func TestRoomDispatchersApplyMutationsAndCancellation(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	scope := pluginruntime.NewScope("rooms")
	name := "before"
	mode := roommodel.DoorModeOpen
	_ = hub.listen(scope, sdkevent.RoomUpdateName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.RoomUpdate)
		replacement := "after"
		replacementMode := int16(roommodel.DoorModePassword)
		event.Params.Name, event.Params.DoorMode = &replacement, &replacementMode
		event.SetCancelled(true)
		return nil
	})
	params, cancelled := hub.DispatchRoomUpdate(context.Background(), testPlayer(), 3, roomservice.UpdateParams{
		Name: &name, DoorMode: &mode, AllowReservedTags: true,
	})
	if !cancelled || params.Name == nil || *params.Name != "after" || params.DoorMode == nil ||
		*params.DoorMode != roommodel.DoorModePassword || !params.AllowReservedTags {
		t.Fatalf("params=%#v cancelled=%v", params, cancelled)
	}
	_ = hub.listen(scope, sdkevent.RoomUnitMoveName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.RoomUnitMove)
		event.TargetX, event.TargetY = 8, 9
		return nil
	})
	x, y, cancelled := hub.DispatchRoomUnitMove(context.Background(), testPlayer(), 3, 1, 2)
	if x != 8 || y != 9 || cancelled {
		t.Fatalf("target=%d,%d cancelled=%v", x, y, cancelled)
	}
	_ = hub.listen(scope, sdkevent.RoomEnterAttemptName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		current.(sdkevent.Cancellable).SetCancelled(true)
		return nil
	})
	if !hub.DispatchRoomEnterAttempt(context.Background(), testPlayer(), 3, true) {
		t.Fatal("room entry was not cancelled")
	}
	_ = hub.listen(scope, sdkevent.RoomCreateName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.RoomCreate)
		event.RoomName, event.ModelName = "plugin room", "model_b"
		return nil
	})
	created, cancelled := hub.DispatchRoomCreate(context.Background(), roomservice.CreateParams{
		OwnerPlayerID: 7, Name: "room", ModelName: "model_a",
	})
	if cancelled || created.Name != "plugin room" || created.ModelName != "model_b" {
		t.Fatalf("created=%#v cancelled=%v", created, cancelled)
	}
}

// BenchmarkRoomDispatchersWithoutListeners measures original event guard allocations.
func BenchmarkRoomDispatchersWithoutListeners(b *testing.B) {
	hub := NewHub(time.Second, zap.NewNop())
	ctx := context.Background()
	player := testPlayer()
	params := roomservice.UpdateParams{}
	b.Run("update", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = hub.DispatchRoomUpdate(ctx, player, 3, params)
		}
	})
	b.Run("unit_move", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _, _ = hub.DispatchRoomUnitMove(ctx, player, 3, 1, 2)
		}
	})
	b.Run("enter_attempt", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = hub.DispatchRoomEnterAttempt(ctx, player, 3, false)
		}
	})
}
