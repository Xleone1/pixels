package event

import (
	"context"
	"errors"
	"time"

	sdkevent "github.com/niflaot/pixels/sdk/event"
)

// DispatchSanctionApply sends one cancellable global sanction request.
func (hub *Hub) DispatchSanctionApply(ctx context.Context, actorID *int64, targetID int64, kind string, reason string, expiresAt *time.Time) (string, *time.Time, bool) {
	var actorIDValue int64
	if actorID != nil {
		actorIDValue = *actorID
	}
	event := &sdkevent.SanctionApply{
		Actor: hub.player(actorIDValue), Target: hub.player(targetID), Kind: kind,
		Reason: reason, ExpiresAt: cloneTime(expiresAt),
	}
	err := hub.Dispatch(ctx, event)
	return event.Reason, cloneTime(event.ExpiresAt), errors.Is(err, ErrEventCancelled)
}

// DispatchRoomModerationAction sends one cancellable local moderation request.
func (hub *Hub) DispatchRoomModerationAction(ctx context.Context, action string, roomID int64, actorID int64, targetID int64) bool {
	event := &sdkevent.RoomModerationAction{
		Action: action, RoomID: roomID, Actor: hub.player(actorID), Target: hub.player(targetID),
	}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// cloneTime copies one optional timestamp.
func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
