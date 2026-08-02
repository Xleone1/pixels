// Package variable owns durable WIRED variable state and its hot cache.
package variable

import (
	"context"
	"time"
)

// Scope identifies a variable target family.
type Scope uint8

const (
	// ScopeFurni stores a value on one furniture item.
	ScopeFurni Scope = iota + 1
	// ScopeUser stores a value for one user in one room.
	ScopeUser
	// ScopeRoom stores a shared room value.
	ScopeRoom
	// ScopeReference stores a cross-room reference value.
	ScopeReference
	// ScopeContext exposes one execution event without durable persistence.
	ScopeContext
)

// Value stores one durable typed variable assignment.
type Value struct {
	// RoomID identifies the owning room.
	RoomID int64
	// Scope identifies the target family.
	Scope Scope
	// ScopeID identifies the furniture, user, room, or reference.
	ScopeID int64
	// Name identifies the variable definition.
	Name string
	// IntValue stores the numeric projection.
	IntValue int64
	// StringValue stores the textual projection.
	StringValue string
	// UpdatedByPlayerID identifies the latest player editor when available.
	UpdatedByPlayerID int64
	// CreatedAt stores the first assignment instant.
	CreatedAt time.Time
	// UpdatedAt stores the latest mutation instant.
	UpdatedAt time.Time
}

// Store persists WIRED variable assignments.
type Store interface {
	// LoadRoom loads every assignment for one room.
	LoadRoom(context.Context, int64) ([]Value, error)
	// Set creates or replaces one assignment atomically.
	Set(context.Context, Value) (Value, error)
	// Delete removes one exact assignment.
	Delete(context.Context, int64, Scope, int64, string) (bool, error)
}

// Finder reads one durable assignment when its room is not warmed locally.
type Finder interface {
	// Find returns one exact persisted assignment.
	Find(context.Context, int64, Scope, int64, string) (Value, bool, error)
}

// SystemProvider resolves immutable variables derived from live room state.
type SystemProvider interface {
	// ResolveSystem returns one exact system variable without persistence.
	ResolveSystem(int64, Scope, int64, string) (Value, bool)
	// ListSystem returns system variables for one exact scope target.
	ListSystem(int64, Scope, int64) []Value
}

// key indexes one variable inside a room cache.
type key struct {
	// scope identifies the variable target family.
	scope Scope
	// scopeID identifies the concrete target.
	scopeID int64
	// name identifies the variable definition.
	name string
}
