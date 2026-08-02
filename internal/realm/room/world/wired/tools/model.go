// Package tools implements the versioned WIRED Creator Tools protocol.
package tools

import (
	"errors"
	"sync"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// SchemaVersion identifies the Creator Tools JSON contract.
	SchemaVersion int32 = 1
	// maximumDocumentBytes keeps one protocol string below the codec limit.
	maximumDocumentBytes = 60 * 1024
	// maximumEntities bounds one snapshot entity list.
	maximumEntities = 512
	// maximumVariables bounds one snapshot variable list.
	maximumVariables = 1024
)

var (
	// ErrForbidden reports missing room or global WIRED authority.
	ErrForbidden = errors.New("WIRED Creator Tools access denied")
	// ErrInvalidRequest reports malformed room, entity, or mutation input.
	ErrInvalidRequest = errors.New("invalid WIRED Creator Tools request")
	// ErrDocumentTooLarge reports a snapshot beyond the bounded packet contract.
	ErrDocumentTooLarge = errors.New("WIRED Creator Tools document too large")
)

var (
	// VariablesManage allows mutable Creator Tools variable operations.
	VariablesManage = permission.RegisterNode("room.wired.variables.manage", "")
)

// Service coordinates Creator Tools authorization, documents, and mutations.
type Service struct {
	// config stores normalized WIRED budgets.
	config roomwired.Config
	// players stores connected player state.
	players *playerlive.Registry
	// bindings resolves authenticated sessions.
	bindings *binding.Registry
	// connections sends panel-open packets to command issuers.
	connections *netconn.Registry
	// rooms stores active room state.
	rooms *roomlive.Registry
	// records reads and updates durable room metadata.
	records roomservice.ConfigManager
	// runtime exposes bounded WIRED telemetry.
	runtime *wiredruntime.Engine
	// variables owns durable and system variable state.
	variables *variable.Service
	// permissions resolves global access overrides.
	permissions permissionservice.Checker
	// preferencesMutex protects session preferences.
	preferencesMutex sync.RWMutex
	// preferences stores process-local panel settings by player and room.
	preferences map[preferenceKey]Preferences
}

// preferenceKey identifies one player's panel settings in one room.
type preferenceKey struct {
	// playerID identifies the connected player.
	playerID int64
	// roomID identifies the active room.
	roomID int64
}

// Preferences stores Creator Tools panel behavior.
type Preferences struct {
	// LiveUpdates enables automatic client refresh requests.
	LiveUpdates bool `json:"liveUpdates"`
	// HideBoxes stores the durable room WIRED visibility setting.
	HideBoxes bool `json:"hideBoxes"`
}

// PermissionsDocument stores effective Creator Tools capabilities.
type PermissionsDocument struct {
	// CanInspect reports read authority.
	CanInspect bool `json:"canInspect"`
	// CanMutate reports variable mutation authority.
	CanMutate bool `json:"canMutate"`
	// CanConfigure reports room WIRED editor authority.
	CanConfigure bool `json:"canConfigure"`
}

// SnapshotDocument stores one bounded Creator Tools room snapshot.
type SnapshotDocument struct {
	// SchemaVersion identifies this document contract.
	SchemaVersion int32 `json:"schemaVersion"`
	// Room stores room identity.
	Room RoomDocument `json:"room"`
	// Revision stores the room metadata revision as an exact decimal value.
	Revision string `json:"revision"`
	// Usage stores WIRED limits and current use.
	Usage UsageDocument `json:"usage"`
	// Events stores the bounded trace history.
	Events []EventDocument `json:"events"`
	// Variables stores durable room variables.
	Variables []VariableDocument `json:"variables"`
	// Entities stores inspectable room entities.
	Entities []EntityDocument `json:"entities"`
	// Permissions stores effective actor capabilities.
	Permissions PermissionsDocument `json:"permissions"`
	// Settings stores panel and room visibility preferences.
	Settings Preferences `json:"settings"`
}

// RoomDocument stores Creator Tools room identity.
type RoomDocument struct {
	// ID identifies the room.
	ID int64 `json:"id"`
	// Name stores the visible room name.
	Name string `json:"name"`
}

// UsageDocument stores current and configured room WIRED budgets.
type UsageDocument struct {
	// WiredFurni stores compiled WIRED furniture count.
	WiredFurni int `json:"wiredFurni"`
	// MaxWiredFurni stores the room compilation ceiling.
	MaxWiredFurni int `json:"maxWiredFurni"`
	// Effects stores compiled effect count.
	Effects int `json:"effects"`
	// LatestTraceEffects stores effects attempted by the latest completed trace.
	LatestTraceEffects int `json:"latestTraceEffects"`
	// MaxEffects stores the per-trace effect budget.
	MaxEffects int `json:"maxEffects"`
	// Stacks stores compiled stack count.
	Stacks int `json:"stacks"`
	// LatestTraceStacks stores stacks visited by the latest completed trace.
	LatestTraceStacks int `json:"latestTraceStacks"`
	// MaxStacks stores the per-trace stack budget.
	MaxStacks int `json:"maxStacks"`
	// Delayed stores outstanding delayed actions.
	Delayed int `json:"delayed"`
	// MaxDelayed stores the room delayed-action budget.
	MaxDelayed int `json:"maxDelayed"`
	// Variables stores durable room variable count.
	Variables int `json:"variables"`
	// MaxVariables stores the room variable budget.
	MaxVariables int `json:"maxVariables"`
	// Signals stores signals retained outside the current trace.
	Signals int `json:"signals"`
	// MaxSignals stores the per-trace signal budget.
	MaxSignals int `json:"maxSignals"`
	// ExecutionMicros stores latest execution duration.
	ExecutionMicros int64 `json:"executionMicros"`
	// CompileFailures stores failed generation compilations observed by this process.
	CompileFailures string `json:"compileFailures"`
}

// EventDocument stores one execution trace row.
type EventDocument struct {
	// ID identifies the source event.
	ID string `json:"id"`
	// Kind identifies the source trigger family.
	Kind string `json:"kind"`
	// OccurredAt stores the RFC3339 timestamp.
	OccurredAt string `json:"occurredAt"`
	// Message stores the concise trace summary.
	Message string `json:"message"`
	// StackID optionally identifies a stack.
	StackID string `json:"stackId,omitempty"`
}

// VariableDocument stores one durable or system variable row.
type VariableDocument struct {
	// Scope stores the semantic target family.
	Scope string `json:"scope"`
	// ScopeID stores the decimal target identifier.
	ScopeID string `json:"scopeId"`
	// ScopeName stores the target display name.
	ScopeName string `json:"scopeName"`
	// Name stores the canonical variable name.
	Name string `json:"name"`
	// Type stores integer or string.
	Type string `json:"type"`
	// IntValue stores the exact decimal integer value.
	IntValue string `json:"intValue"`
	// StringValue stores the textual value.
	StringValue string `json:"stringValue"`
	// CreatorName stores the author when known.
	CreatorName string `json:"creatorName"`
	// UpdatedAt stores the RFC3339 timestamp.
	UpdatedAt string `json:"updatedAt"`
	// Revision stores an opaque decimal revision.
	Revision string `json:"revision"`
	// ReadOnly reports immutable system state.
	ReadOnly bool `json:"readOnly"`
	// System reports a derived variable.
	System bool `json:"system"`
}

// EntityDocument stores one inspectable room entity.
type EntityDocument struct {
	// Type identifies furni or user.
	Type string `json:"type"`
	// ID stores the exact decimal entity identifier.
	ID string `json:"id"`
	// Name stores the visible entity name.
	Name string `json:"name"`
}

// InspectionDocument stores one entity and its variables.
type InspectionDocument struct {
	// SchemaVersion identifies this document contract.
	SchemaVersion int32 `json:"schemaVersion"`
	// Revision stores the latest exact variable revision.
	Revision string `json:"revision"`
	// Entity stores inspected identity.
	Entity EntityDocument `json:"entity"`
	// Variables stores durable and system values.
	Variables []VariableDocument `json:"variables"`
	// Permissions stores effective actor capabilities.
	Permissions PermissionsDocument `json:"permissions"`
}
