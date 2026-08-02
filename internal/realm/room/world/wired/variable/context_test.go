package variable

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestResolveContext verifies numeric, textual, and unknown event variables.
func TestResolveContext(t *testing.T) {
	now := time.Unix(100, 0)
	event := trigger.Event{ID: 9, RoomID: 4, PlayerID: 7, Message: "hello", ActorIDs: []int64{7, 8}}
	cases := []struct {
		name  string
		value int64
		text  string
		found bool
	}{
		{name: "@context.player_id", value: 7, found: true},
		{name: "@context.message", text: "hello", found: true},
		{name: "@context.actor_count", value: 2, found: true},
		{name: "@context.missing", found: false},
	}
	for _, current := range cases {
		value, found := ResolveContext(event, current.name, now)
		if found != current.found || value.IntValue != current.value || value.StringValue != current.text {
			t.Fatalf("ResolveContext(%q) = %#v, %v", current.name, value, found)
		}
	}
}

// TestListContext verifies the stable bounded context catalog.
func TestListContext(t *testing.T) {
	values := ListContext(trigger.Event{ID: 1, RoomID: 2}, time.Unix(3, 0))
	if len(values) != len(contextNames) {
		t.Fatalf("context values = %d, want %d", len(values), len(contextNames))
	}
	for _, value := range values {
		if value.Scope != ScopeContext || value.ScopeID != 1 || value.RoomID != 2 {
			t.Fatalf("unexpected context value %#v", value)
		}
	}
}
