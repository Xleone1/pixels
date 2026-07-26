// Package identity owns player username availability and committed renames.
package identity

import (
	"errors"
	"time"
)

const (
	// ResultAvailable reports an available valid username.
	ResultAvailable int32 = 0
	// ResultTooShort reports a username below the minimum length.
	ResultTooShort int32 = 2
	// ResultTooLong reports a username above the maximum length.
	ResultTooLong int32 = 3
	// ResultInvalid reports unsupported username characters.
	ResultInvalid int32 = 4
	// ResultTaken reports a reserved or existing username.
	ResultTaken int32 = 5
	// ResultCooldown reports that the username cooldown is still active.
	ResultCooldown int32 = 6
)

var (
	// ErrReservationMissing reports a commit without a matching short reservation.
	ErrReservationMissing = errors.New("username reservation missing")
	// ErrRenameCooldown reports that the durable rename cooldown is active.
	ErrRenameCooldown = errors.New("username change cooldown active")
	// ErrUsernameTaken reports a database uniqueness conflict.
	ErrUsernameTaken = errors.New("username taken")
	// ErrPlayerNotFound reports a missing active target player.
	ErrPlayerNotFound = errors.New("player not found")
	// ErrInvalidAttribution reports incomplete administrative audit data.
	ErrInvalidAttribution = errors.New("invalid name change attribution")
)

// CheckResult contains username policy and availability output.
type CheckResult struct {
	// Code is the Nitro result code.
	Code int32
	// Username stores the normalized candidate.
	Username string
	// Suggestions stores bounded available alternatives.
	Suggestions []string
	// AvailableAt stores the next eligible instant when the cooldown is active.
	AvailableAt *time.Time
}

// RenameResult contains one committed identity replacement.
type RenameResult struct {
	// OldUsername stores the prior visible name.
	OldUsername string
	// NewUsername stores the committed visible name.
	NewUsername string
	// ChangedAt stores the committed mutation instant.
	ChangedAt time.Time
	// AvailableAt stores the next self-service eligibility instant.
	AvailableAt time.Time
}

// NameChangeStatus describes the current automatic cooldown.
type NameChangeStatus struct {
	// Available reports whether a self-service rename may start now.
	Available bool
	// AvailableAt stores the next eligible instant when currently unavailable.
	AvailableAt *time.Time
	// LastChangedAt stores the latest committed username replacement.
	LastChangedAt *time.Time
	// RemainingSeconds stores the rounded-up cooldown remainder.
	RemainingSeconds int64
	// CooldownDays stores the configured interval in full days.
	CooldownDays int
}

// NameChange contains one durable username audit entry.
type NameChange struct {
	// ID identifies the audit entry.
	ID int64
	// PlayerID identifies the renamed player.
	PlayerID int64
	// OldUsername stores the previous visible name.
	OldUsername string
	// NewUsername stores the resulting visible name.
	NewUsername string
	// ActorPlayerID identifies the player responsible for the change.
	ActorPlayerID int64
	// Reason explains why the name was changed.
	Reason string
	// Source identifies the mutation boundary.
	Source string
	// ChangedAt stores the committed mutation time.
	ChangedAt time.Time
}
