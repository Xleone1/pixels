package essential

import (
	"context"
	"testing"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRemoteSwitchTogglesExactlyOnce verifies rights-based immediate toggling.
func TestRemoteSwitchTogglesExactlyOnce(t *testing.T) {
	item := essentialItem("switch_remote_control", 3)
	active := essentialRoom(t, item, 1)
	states := &stateRecorder{}
	service := &Service{states: states}
	if err := service.useTraversal(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use remote switch: %v", err)
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "1" || len(states.params) != 1 {
		t.Fatalf("expected one toggle state=%q writes=%d", updated.ExtraData, len(states.params))
	}
}

// TestOneWayGateCrossesAndCloses verifies the controlled two-step traversal.
func TestOneWayGateCrossesAndCloses(t *testing.T) {
	item := essentialItem("onewaygate", 2)
	item.Point = grid.MustPoint(2, 0)
	active := essentialRoom(t, item, 1)
	if _, err := active.TeleportUnit(1, grid.MustPoint(3, 0), worldunit.RotationWest, false); err != nil {
		t.Fatalf("position actor: %v", err)
	}
	service := &Service{}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useTraversal(context.Background(), request); err != nil {
		t.Fatalf("use one-way gate: %v", err)
	}
	active.Tick()
	active.RunScheduled(time.Now().Add(time.Second))
	active.Tick()
	active.RunScheduled(time.Now().Add(2 * time.Second))
	unit, _ := active.Unit(1)
	closed, _ := active.FurnitureItem(item.ID)
	if unit.Position.Point != grid.MustPoint(1, 0) || closed.ExtraData != "0" {
		t.Fatalf("unexpected crossing point=%v state=%q", unit.Position.Point, closed.ExtraData)
	}
}

// TestMultiheightRebuildsPhysicalState verifies durable state and resolved height.
func TestMultiheightRebuildsPhysicalState(t *testing.T) {
	item := essentialItem("multiheight", 3)
	item.Definition.Multiheight = "0;1;2"
	item.Definition.AllowWalk = true
	item.Definition.StackHeight = 0
	active := essentialRoom(t, item, 1)
	states := &stateRecorder{}
	service := &Service{states: states}
	if err := service.useTraversal(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use multiheight: %v", err)
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "1" || updated.Definition.StackHeight != grid.HeightFromInt(1) {
		t.Fatalf("unexpected multiheight state=%q height=%d", updated.ExtraData, updated.Definition.StackHeight)
	}
}

// TestSwitchWalksToNearestActivator verifies delayed toggling after controlled movement.
func TestSwitchWalksToNearestActivator(t *testing.T) {
	item := essentialItem("switch", 2)
	item.Point = grid.MustPoint(4, 0)
	active := essentialRoom(t, item, 1)
	states := &stateRecorder{}
	service := &Service{states: states}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useTraversal(context.Background(), request); err != nil {
		t.Fatalf("start switch walk: %v", err)
	}
	for range 8 {
		active.Tick()
		active.RunScheduled(time.Now().Add(time.Second))
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "1" || len(states.params) != 1 {
		t.Fatalf("expected one delayed toggle state=%q writes=%d", updated.ExtraData, len(states.params))
	}
}

// TestOneWayGateRequiresFrontActivation verifies side clicks do not start traversal.
func TestOneWayGateRequiresFrontActivation(t *testing.T) {
	item := essentialItem("onewaygate", 2)
	item.Point = grid.MustPoint(3, 0)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	if err := service.useTraversal(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use gate from invalid tile: %v", err)
	}
	unit, _ := active.Unit(1)
	if unit.Position.Point != grid.MustPoint(0, 0) || unit.Moving {
		t.Fatalf("unexpected gate movement %#v", unit)
	}
}

// TestMultiheightRejectsSingleState verifies immutable definitions remain unchanged.
func TestMultiheightRejectsSingleState(t *testing.T) {
	item := essentialItem("multiheight", 1)
	item.Definition.Multiheight = "1"
	active := essentialRoom(t, item, 1)
	service := &Service{states: &stateRecorder{}}
	if err := service.useTraversal(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use single-height item: %v", err)
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "0" {
		t.Fatalf("unexpected state %q", updated.ExtraData)
	}
}

// TestPostureMultiheightSeparatesUseFromHeightAuthority verifies every player can sit while only room managers change the physical state.
func TestPostureMultiheightSeparatesUseFromHeightAuthority(t *testing.T) {
	for _, test := range []struct {
		name       string
		playerID   int64
		wantState  string
		wantHeight grid.Height
		wantWrites int
	}{
		{name: "owner changes height", playerID: 1, wantState: "1", wantHeight: grid.HeightFromUnits(1.5), wantWrites: 1},
		{name: "guest only sits", playerID: 2, wantState: "0", wantHeight: grid.HeightFromUnits(0.5)},
	} {
		t.Run(test.name, func(t *testing.T) {
			item := essentialItem("multiheight", 2)
			item.Definition.AllowSit = true
			item.Definition.Multiheight = "0.50;1.50"
			item.Definition.StackHeight = grid.HeightFromUnits(0.5)
			item.Definition.Slots = []worldfurniture.SlotDefinition{{
				Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth,
			}}
			active := essentialRoom(t, item, test.playerID)
			states := &stateRecorder{}
			service := &Service{states: states}
			handled, err := service.Use(context.Background(), Request{PlayerID: test.playerID, Room: active, Item: item})
			if err != nil || !handled {
				t.Fatalf("use posture multiheight handled=%t err=%v", handled, err)
			}
			updated, found := active.FurnitureItem(item.ID)
			if !found || updated.ExtraData != test.wantState || updated.Definition.StackHeight != test.wantHeight {
				t.Fatalf("updated=%#v found=%t", updated, found)
			}
			if len(states.params) != test.wantWrites {
				t.Fatalf("state writes=%d want=%d", len(states.params), test.wantWrites)
			}
		})
	}
}
