package variable

import (
	"strings"
	"sync"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

var contextNames = [...]string{
	"@context.event_id", "@context.kind", "@context.actor_id",
	"@context.player_id", "@context.target_player_id", "@context.username",
	"@context.source_item", "@context.source_sprite", "@context.message",
	"@context.score", "@context.previous_score", "@context.team",
	"@context.action", "@context.action_value", "@context.direction",
	"@context.signal", "@context.counter", "@context.previous_counter",
	"@context.variable_name", "@context.variable_value",
	"@context.previous_variable_value", "@context.reference_room_id",
	"@context.actor_count", "@context.furni_count",
}

// ResolveContext resolves one immutable variable from the current execution event.
func ResolveContext(event trigger.Event, name string, now time.Time) (Value, bool) {
	name = normalize(name)
	value := Value{
		RoomID: event.RoomID, Scope: ScopeContext, ScopeID: contextScopeID(event),
		Name: name, CreatedAt: now, UpdatedAt: now,
	}
	switch name {
	case "@context.event_id":
		value.IntValue = int64(event.ID)
	case "@context.kind":
		value.IntValue = int64(event.Kind)
	case "@context.actor_id":
		value.IntValue = event.ActorID
	case "@context.player_id":
		value.IntValue = event.PlayerID
	case "@context.target_player_id":
		value.IntValue = event.TargetPlayerID
	case "@context.username":
		value.StringValue = event.Username
	case "@context.source_item":
		value.IntValue = event.SourceItem
	case "@context.source_sprite":
		value.IntValue = int64(event.SourceSprite)
	case "@context.message":
		value.StringValue = event.Message
	case "@context.score":
		value.IntValue = event.Score
	case "@context.previous_score":
		value.IntValue = event.PreviousScore
	case "@context.team":
		value.IntValue = int64(event.Team)
	case "@context.action":
		value.IntValue = int64(event.Action)
	case "@context.action_value":
		value.IntValue = int64(event.ActionValue)
	case "@context.direction":
		value.IntValue = int64(event.Direction)
	case "@context.signal":
		value.StringValue = event.Signal
	case "@context.counter":
		value.IntValue = event.Counter
	case "@context.previous_counter":
		value.IntValue = event.PreviousCounter
	case "@context.variable_name":
		value.StringValue = event.VariableName
	case "@context.variable_value":
		value.IntValue = event.VariableValue
	case "@context.previous_variable_value":
		value.IntValue = event.PreviousVariableValue
	case "@context.reference_room_id":
		value.IntValue = event.ReferenceRoomID
	case "@context.actor_count":
		value.IntValue = int64(len(event.ActorIDs))
	case "@context.furni_count":
		value.IntValue = int64(len(event.FurniTargets))
	default:
		return Value{}, false
	}
	return value, true
}

// ListContext returns the bounded immutable variable set for one execution event.
func ListContext(event trigger.Event, now time.Time) []Value {
	values := make([]Value, 0, len(contextNames))
	for _, name := range contextNames {
		value, _ := ResolveContext(event, name, now)
		values = append(values, value)
	}
	return values
}

// contextScopeID returns a nonzero event-local scope identifier.
func contextScopeID(event trigger.Event) int64 {
	if event.ID > 0 {
		return int64(event.ID)
	}
	return event.RoomID
}

// writeLock returns the stable mutation stripe for one room.
func (service *Service) writeLock(roomID int64) *sync.Mutex {
	index := uint64(roomID) % uint64(len(service.writes))
	return &service.writes[index]
}

// normalize creates the canonical case-insensitive name.
func normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

// valid reports whether one assignment can be persisted.
func valid(value Value) bool {
	return value.RoomID > 0 && value.Scope >= ScopeFurni && value.Scope <= ScopeReference &&
		value.ScopeID > 0 && value.Name != "" && len(value.Name) <= 64 &&
		len(value.StringValue) <= 2048 && !value.UpdatedAt.After(time.Now().Add(time.Minute))
}
