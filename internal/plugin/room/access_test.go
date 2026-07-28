package room

import (
	"context"
	"errors"
	"testing"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	muteallchanged "github.com/niflaot/pixels/internal/realm/room/control/events/muteallchanged"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

// roomStore implements the persistence calls exercised by plugin room access.
type roomStore struct {
	// Store supplies unused persistence methods.
	roomservice.Store
	// room stores the durable fixture.
	room roommodel.Room
	// found controls room lookup and optimistic update results.
	found bool
	// findErr configures lookup failure.
	findErr error
	// updateErr configures update failure.
	updateErr error
}

// FindRoomByID returns the configured durable fixture.
func (store *roomStore) FindRoomByID(context.Context, int64) (roommodel.Room, bool, error) {
	return store.room, store.found, store.findErr
}

// UpdateRoom records one complete durable settings update.
func (store *roomStore) UpdateRoom(_ context.Context, params roomservice.UpdateRecordParams, _ []string) (roommodel.Room, bool, error) {
	if store.updateErr != nil {
		return roommodel.Room{}, false, store.updateErr
	}
	store.room = params.Room
	store.room.Version.Version++
	return store.room, store.found, nil
}

// ListCategories returns the room fixture category when configured.
func (store *roomStore) ListCategories(context.Context) ([]roommodel.Category, error) {
	if store.room.CategoryID == nil {
		return nil, nil
	}
	return []roommodel.Category{{
		Base:    sharedmodel.Base{Identity: sharedmodel.Identity{ID: *store.room.CategoryID}},
		Visible: true,
	}}, nil
}

// TestSetMuteAllProjectsStateAndCommittedFact verifies the room action follows native projections.
func TestSetMuteAllProjectsStateAndCommittedFact(t *testing.T) {
	local := bus.New()
	registry := roomlive.NewRegistry(nil)
	active, err := registry.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = registry.Close(context.Background(), 9) })
	if _, err = registry.Join(context.Background(), 9, roomlive.Occupant{
		PlayerID: 7, Username: "demo", ConnectionID: netconn.ID("test"), ConnectionKind: "websocket",
	}); err != nil {
		t.Fatal(err)
	}
	published := 0
	_, _ = local.Subscribe(muteallchanged.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		payload, valid := event.Payload.(muteallchanged.Payload)
		if valid && payload.RoomID == 9 && payload.Muted {
			published++
		}
		return nil
	})
	access := NewAccess(nil, registry, nil, local, pluginruntime.NewScope("rooms"))
	if err := access.SetMuteAll(9, true); err != nil {
		t.Fatal(err)
	}
	if !active.MuteAll() || published != 1 {
		t.Fatalf("muted=%v published=%d", active.MuteAll(), published)
	}
	if occupants := access.Occupants(9); len(occupants) != 1 || occupants[0] != 7 {
		t.Fatalf("occupants=%v", occupants)
	}
}

// TestRoomAccessFailsClosedOutsideEnabledScope verifies late plugin work cannot mutate rooms.
func TestRoomAccessFailsClosedOutsideEnabledScope(t *testing.T) {
	scope := pluginruntime.NewScope("disabled")
	scope.Disable()
	access := NewAccess(nil, roomlive.NewRegistry(nil), nil, nil, scope)
	if err := access.SetMuteAll(9, true); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if _, found := access.Find(9); found {
		t.Fatal("disabled scope read a room")
	}
	if occupants := access.Occupants(9); occupants != nil {
		t.Fatalf("disabled occupants=%v", occupants)
	}
	if _, err := access.Update(9, sdkplugin.RoomUpdateParams{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("disabled update error=%v", err)
	}
}

// TestRoomAccessReadsAndUpdatesDetachedSnapshots verifies durable room behavior.
func TestRoomAccessReadsAndUpdatesDetachedSnapshots(t *testing.T) {
	categoryID := int64(4)
	store := &roomStore{
		found: true,
		room: roommodel.Room{
			Base:             sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}, Version: sharedmodel.Version{Version: 3}},
			OwnerPlayerID:    7,
			OwnerName:        "demo",
			Name:             "Original Room",
			Description:      "description",
			ModelName:        "model_a",
			DoorMode:         roommodel.DoorModeOpen,
			MaxUsers:         25,
			Score:            8,
			CategoryID:       &categoryID,
			TradeMode:        roommodel.TradeModeAllowed,
			RollerSpeed:      2,
			AllowWalkthrough: true,
			AllowPets:        true,
			AllowPetsEat:     true,
			HideWalls:        true,
			HideWired:        true,
			WallThickness:    -1,
			FloorThickness:   1,
			StaffPicked:      true,
			PublicRoom:       true,
		},
	}
	service := roomservice.New(store, nil)
	access := NewAccess(service, roomlive.NewRegistry(nil), nil, nil, pluginruntime.NewScope("rooms"))
	current, found := access.Find(9)
	if !found || current.ID != 9 || current.Version != 3 || current.CategoryID == nil || *current.CategoryID != 4 {
		t.Fatalf("current=%#v found=%v", current, found)
	}
	*current.CategoryID = 99
	if *store.room.CategoryID != 4 {
		t.Fatal("snapshot leaked its category pointer")
	}
	name := "Updated Room"
	tags := []string{"builders"}
	mode := int16(roommodel.DoorModeDoorbell)
	trade := int16(roommodel.TradeModeController)
	moderation := int16(roommodel.ModerationPolicyOwnerAndRights)
	maxUsers := 30
	updated, err := access.Update(9, sdkevent.RoomUpdateParams{
		Name: &name, Tags: &tags, MaxUsers: &maxUsers, DoorMode: &mode, TradeMode: &trade,
		ModerationMute: &moderation, ModerationKick: &moderation, ModerationBan: &moderation,
	})
	if err != nil || updated.Name != name || updated.Version != 4 || updated.MaxUsers != 30 ||
		updated.DoorMode != mode || updated.TradeMode != trade {
		t.Fatalf("updated=%#v err=%v", updated, err)
	}
}

// TestRoomAccessPropagatesLookupAndPersistenceFailures verifies failure boundaries.
func TestRoomAccessPropagatesLookupAndPersistenceFailures(t *testing.T) {
	sentinel := errors.New("room persistence failed")
	scope := pluginruntime.NewScope("rooms")
	tests := []struct {
		// name describes the persistence outcome.
		name string
		// store configures the persistence fixture.
		store *roomStore
		// expected identifies the returned error.
		expected error
	}{
		{name: "lookup", store: &roomStore{findErr: sentinel}, expected: sentinel},
		{name: "missing", store: &roomStore{}, expected: roomservice.ErrRoomNotFound},
		{
			name: "update",
			store: &roomStore{
				found: true,
				room: roommodel.Room{
					Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}, Version: sharedmodel.Version{Version: 1}},
					Name: "Valid Room", MaxUsers: 25,
				},
				updateErr: sentinel,
			},
			expected: sentinel,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tags := []string{}
			_, err := NewAccess(roomservice.New(testCase.store, nil), nil, nil, nil, scope).Update(9, sdkevent.RoomUpdateParams{Tags: &tags})
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
		})
	}
	if _, found := NewAccess(roomservice.New(&roomStore{findErr: sentinel}, nil), nil, nil, nil, scope).Find(9); found {
		t.Fatal("failed lookup returned a room")
	}
	if err := NewAccess(nil, roomlive.NewRegistry(nil), nil, nil, scope).SetMuteAll(99, true); !errors.Is(err, ErrInactiveRoom) {
		t.Fatalf("inactive mute error=%v", err)
	}
}

// TestRoomConversionHelpersPreserveOptionalValues verifies complete SDK conversion.
func TestRoomConversionHelpersPreserveOptionalValues(t *testing.T) {
	value := int16(2)
	converted := fromSDK(sdkevent.RoomUpdateParams{
		DoorMode: &value, TradeMode: &value, ModerationMute: &value, ModerationKick: &value, ModerationBan: &value,
	})
	if converted.DoorMode == nil || int16(*converted.DoorMode) != value || converted.TradeMode == nil ||
		converted.ModerationMute == nil || converted.ModerationKick == nil || converted.ModerationBan == nil {
		t.Fatalf("converted=%#v", converted)
	}
	if typedPointer[roommodel.DoorMode](nil) != nil || cloneInt64(nil) != nil {
		t.Fatal("nil option was materialized")
	}
}
