package room

import (
	"testing"

	wiredvariable "github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// TestSystemVariablesResolveLiveRoomEntities verifies room, furniture, and user projections.
func TestSystemVariablesResolveLiveRoomEntities(t *testing.T) {
	rooms, active := conditionRoom(t)
	provider := New(rooms, nil, nil, nil)
	roomID := active.ID()
	if value, found := provider.ResolveSystem(roomID, wiredvariable.ScopeRoom, roomID, "@dimensions.x"); !found || value.IntValue != 3 {
		t.Fatalf("room dimensions=%+v found=%t", value, found)
	}
	if value, found := provider.ResolveSystem(roomID, wiredvariable.ScopeFurni, 10, "@can_stand_on"); !found || value.IntValue != 1 {
		t.Fatalf("furniture walkability=%+v found=%t", value, found)
	}
	if value, found := provider.ResolveSystem(roomID, wiredvariable.ScopeFurni, 10, "@state"); !found || value.StringValue != "1" {
		t.Fatalf("furniture state=%+v found=%t", value, found)
	}
	if value, found := provider.ResolveSystem(roomID, wiredvariable.ScopeUser, 7, "@username"); !found || value.StringValue != "demo" {
		t.Fatalf("username=%+v found=%t", value, found)
	}
	if values := provider.ListSystem(roomID, wiredvariable.ScopeUser, 7); len(values) != len(userSystemNames) {
		t.Fatalf("user system values=%d want=%d", len(values), len(userSystemNames))
	}
}
