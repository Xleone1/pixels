package core

import (
	"context"
	"net/url"
	"strings"

	"github.com/niflaot/pixels/internal/realm/navigator/record"
)

// UpsertLiftedRoom validates and persists one room's Navigator media.
func (service *Service) UpsertLiftedRoom(ctx context.Context, mutation record.LiftedRoomMutation) (record.LiftedRoom, error) {
	mutation.Image = strings.TrimSpace(mutation.Image)
	mutation.AssetRef = strings.TrimSpace(mutation.AssetRef)
	mutation.Caption = strings.TrimSpace(mutation.Caption)
	if !validLiftedMutation(mutation) {
		return record.LiftedRoom{}, ErrInvalidLiftedRoom
	}
	return service.store.UpsertLiftedRoom(ctx, mutation)
}

// DisableLiftedRoom soft deletes one active Navigator media row.
func (service *Service) DisableLiftedRoom(ctx context.Context, roomID int64, expectedVersion int64, actorPlayerID int64) (record.LiftedRoom, error) {
	if roomID <= 0 || expectedVersion <= 0 || actorPlayerID <= 0 {
		return record.LiftedRoom{}, ErrInvalidLiftedRoom
	}
	return service.store.DisableLiftedRoom(ctx, roomID, expectedVersion, actorPlayerID)
}

// validLiftedMutation reports whether one media mutation is bounded and public.
func validLiftedMutation(mutation record.LiftedRoomMutation) bool {
	if mutation.RoomID <= 0 || mutation.ActorPlayerID <= 0 || mutation.AssetRef == "" || len(mutation.AssetRef) > 255 || mutation.AreaID < 0 || mutation.Order < 0 || len(mutation.Caption) > 255 || len(mutation.Image) > 2048 || mutation.ExpectedVersion < 0 {
		return false
	}
	if mutation.StartsAt != nil && mutation.EndsAt != nil && !mutation.EndsAt.After(*mutation.StartsAt) {
		return false
	}
	parsed, err := url.ParseRequestURI(mutation.Image)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}
