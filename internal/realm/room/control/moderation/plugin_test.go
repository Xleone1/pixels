package moderation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// moderationPluginVeto records and rejects every authorized action.
type moderationPluginVeto struct {
	// actions stores intercepted action names.
	actions []string
}

// DispatchRoomModerationAction records and vetoes one moderation action.
func (events *moderationPluginVeto) DispatchRoomModerationAction(_ context.Context, action string, _ int64, _ int64, _ int64) bool {
	events.actions = append(events.actions, action)
	return true
}

// moderationPublisherForTest records committed moderation events.
type moderationPublisherForTest struct {
	// events stores published facts.
	events []bus.Event
}

// Publish records one committed event.
func (publisher *moderationPublisherForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestPluginVetoPreventsEveryRoomModerationMutation verifies all five local actions stop before effects.
func TestPluginVetoPreventsEveryRoomModerationMutation(t *testing.T) {
	store := &moderationStoreForTest{
		mutes: map[int64]time.Time{2: time.Unix(2000, 0)},
		bans:  map[int64]time.Time{2: time.Unix(2000, 0)},
	}
	nodes := Nodes{OwnKick: "own.kick", OwnMute: "own.mute", OwnBan: "own.ban"}
	permissions := moderationPermissionsForTest{1: {
		permission.Node(nodes.OwnKick): true,
		permission.Node(nodes.OwnMute): true,
		permission.Node(nodes.OwnBan):  true,
	}}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	publisher := &moderationPublisherForTest{}
	service := New(Config{MinMuteMinutes: 1, MaxMuteMinutes: 60}, store, moderationRoomForTest{room: room}, nil, permissions, publisher, nodes)
	service.now = func() time.Time { return time.Unix(1000, 0) }
	veto := &moderationPluginVeto{}
	service.SetPluginRuntime(veto)

	actions := []func() error{
		func() error { return service.Kick(context.Background(), 9, 1, 2) },
		func() error { return service.Mute(context.Background(), 9, 1, 2, 5) },
		func() error { return service.Unmute(context.Background(), 9, 1, 2) },
		func() error { return service.Ban(context.Background(), 9, 1, 2, moderationmodel.BanDurationHour) },
		func() error { return service.Unban(context.Background(), 9, 1, 2) },
	}
	for _, action := range actions {
		if err := action(); !errors.Is(err, ErrCancelledByPlugin) {
			t.Fatalf("expected plugin veto, got %v", err)
		}
	}
	if len(veto.actions) != len(actions) || len(publisher.events) != 0 {
		t.Fatalf("actions=%v events=%v", veto.actions, publisher.events)
	}
	if !store.mutes[2].Equal(time.Unix(2000, 0)) || !store.bans[2].Equal(time.Unix(2000, 0)) {
		t.Fatalf("moderation state changed mutes=%v bans=%v", store.mutes, store.bans)
	}
}
