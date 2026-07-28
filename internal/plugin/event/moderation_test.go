package event

import (
	"context"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	"go.uber.org/zap"
)

// TestModerationDispatchersApplyMutationsAndCancellation verifies both moderation gates.
func TestModerationDispatchersApplyMutationsAndCancellation(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	hub.SetPlayerFinder(playerFinder{found: true})
	scope := pluginruntime.NewScope("moderation")
	replacement := time.Unix(200, 0).UTC()
	_ = hub.listen(scope, sdkevent.SanctionApplyName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.SanctionApply)
		event.Reason, event.ExpiresAt = "normalized", &replacement
		event.SetCancelled(true)
		return nil
	})
	reason, expiresAt, cancelled := hub.DispatchSanctionApply(context.Background(), nil, 7, "warn", "raw", nil)
	if reason != "normalized" || expiresAt == nil || !expiresAt.Equal(replacement) || !cancelled {
		t.Fatalf("reason=%q expires=%v cancelled=%v", reason, expiresAt, cancelled)
	}
	_ = hub.listen(scope, sdkevent.RoomModerationActionName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		current.(sdkevent.Cancellable).SetCancelled(true)
		return nil
	})
	if !hub.DispatchRoomModerationAction(context.Background(), "kick", 3, 7, 8) {
		t.Fatal("room moderation was not cancelled")
	}
}
