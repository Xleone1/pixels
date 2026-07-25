package identity

import (
	"context"
	"strings"
	"unicode/utf8"
)

// RenameAdmin validates and commits one attributed administrative rename.
func (service *Service) RenameAdmin(ctx context.Context, playerID int64, candidate string, actorPlayerID int64, reason string) (RenameResult, error) {
	candidate = strings.TrimSpace(candidate)
	reason = strings.TrimSpace(reason)
	if playerID <= 0 || actorPlayerID <= 0 || reason == "" || utf8.RuneCountInString(reason) > 500 {
		return RenameResult{}, ErrInvalidAttribution
	}
	if service.validate(candidate) != ResultAvailable || service.reserved(candidate) || service.filtered(candidate) {
		return RenameResult{}, ErrReservationMissing
	}
	_, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return RenameResult{}, err
	}
	if !found {
		return RenameResult{}, ErrPlayerNotFound
	}
	if err = service.validateActor(ctx, playerID, actorPlayerID); err != nil {
		return RenameResult{}, err
	}
	existing, taken, err := service.players.FindByUsername(ctx, candidate)
	if err != nil {
		return RenameResult{}, err
	}
	if taken && existing.Player.ID != playerID {
		return RenameResult{}, ErrUsernameTaken
	}
	return service.store.RenameAdmin(ctx, playerID, candidate, actorPlayerID, reason)
}

// SetAuthorization replaces and audits one self-service rename authorization.
func (service *Service) SetAuthorization(ctx context.Context, playerID int64, allowed bool, actorPlayerID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if playerID <= 0 || actorPlayerID <= 0 || reason == "" || utf8.RuneCountInString(reason) > 500 {
		return ErrInvalidAttribution
	}
	_, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return err
	}
	if !found {
		return ErrPlayerNotFound
	}
	if err = service.validateActor(ctx, playerID, actorPlayerID); err != nil {
		return err
	}
	return service.store.SetAuthorization(ctx, playerID, allowed, actorPlayerID, reason)
}

// NameChanges returns a bounded recent username history.
func (service *Service) NameChanges(ctx context.Context, playerID int64, limit int) ([]NameChange, error) {
	if playerID <= 0 {
		return nil, ErrPlayerNotFound
	}
	_, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrPlayerNotFound
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return service.store.NameChanges(ctx, playerID, limit)
}

// validateActor verifies one distinct administrative actor.
func (service *Service) validateActor(ctx context.Context, playerID int64, actorPlayerID int64) error {
	if actorPlayerID == playerID {
		return nil
	}
	_, found, err := service.players.FindByID(ctx, actorPlayerID)
	if err != nil {
		return err
	}
	if !found {
		return ErrInvalidAttribution
	}
	return nil
}
