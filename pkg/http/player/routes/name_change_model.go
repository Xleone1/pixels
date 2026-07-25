package routes

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
)

// NameChangeAuthorizationRequest contains one attributed policy mutation.
type NameChangeAuthorizationRequest struct {
	// Allowed reports whether one self-service rename is enabled.
	Allowed bool `json:"allowed"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the policy mutation.
	Reason string `json:"reason"`
}

// AdminNameChangeRequest contains one attributed direct rename.
type AdminNameChangeRequest struct {
	// Username stores the requested visible name.
	Username string `json:"username"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the identity mutation.
	Reason string `json:"reason"`
}

// SelfNameChangeRequest contains one self-service username candidate.
type SelfNameChangeRequest struct {
	// Username stores the requested visible name.
	Username string `json:"username"`
}

// NameCheckResponse contains one Nitro-compatible username decision.
type NameCheckResponse struct {
	// Code stores the stable username result code.
	Code int32 `json:"code"`
	// Username stores the normalized candidate.
	Username string `json:"username"`
	// Suggestions stores bounded available alternatives.
	Suggestions []string `json:"suggestions"`
}

// NameChangeResponse contains one durable identity audit entry.
type NameChangeResponse struct {
	// ID identifies the audit entry.
	ID int64 `json:"id"`
	// PlayerID identifies the renamed player.
	PlayerID int64 `json:"playerId"`
	// OldUsername stores the prior visible name.
	OldUsername string `json:"oldUsername"`
	// NewUsername stores the resulting visible name.
	NewUsername string `json:"newUsername"`
	// ActorPlayerID identifies the responsible player.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
	// Source identifies the mutation boundary.
	Source string `json:"source"`
	// ChangedAt stores the committed mutation time.
	ChangedAt time.Time `json:"changedAt"`
}

// nameChangeLimit parses one optional bounded history limit.
func nameChangeLimit(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 50, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 || limit > 200 {
		return 0, errors.New("invalid name-change history limit")
	}
	return limit, nil
}

// nameChangeResponse maps one domain audit entry.
func nameChangeResponse(change playeridentity.NameChange) NameChangeResponse {
	return NameChangeResponse{ID: change.ID, PlayerID: change.PlayerID, OldUsername: change.OldUsername,
		NewUsername: change.NewUsername, ActorPlayerID: change.ActorPlayerID, Reason: change.Reason,
		Source: change.Source, ChangedAt: change.ChangedAt}
}

// nameCheckResponse maps one domain username decision.
func nameCheckResponse(result playeridentity.CheckResult) NameCheckResponse {
	suggestions := result.Suggestions
	if suggestions == nil {
		suggestions = []string{}
	}
	return NameCheckResponse{
		Code:        result.Code,
		Username:    result.Username,
		Suggestions: suggestions,
	}
}

// identityError maps identity failures to stable HTTP errors.
func identityError(err error) error {
	switch {
	case errors.Is(err, playeridentity.ErrReservationMissing), errors.Is(err, playeridentity.ErrInvalidAttribution):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, playeridentity.ErrUsernameTaken):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, playeridentity.ErrPlayerNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	default:
		return err
	}
}
