package live

import (
	"testing"
	"time"
)

// TestProfileReportsTickMetrics verifies room-loop observations remain bounded.
func TestProfileReportsTickMetrics(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.recordTick(time.Now().Add(-2*time.Millisecond), true)
	profile := room.Profile()
	if !profile.Available || profile.TickCount != 1 || profile.TickErrors != 1 || profile.LastTickMicroseconds <= 0 || profile.MaxTickMicroseconds <= 0 {
		t.Fatalf("unexpected profile %#v", profile)
	}
}
