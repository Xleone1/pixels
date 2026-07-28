package trade

import (
	"errors"
	"testing"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
)

// tradeAccessServiceForTest creates a real live-trade service around an in-memory registry.
func tradeAccessServiceForTest() (*tradecore.Service, *traderuntime.Registry) {
	registry := traderuntime.NewRegistry()
	service := tradecore.New(tradecore.Options{}, registry, nil, roomlive.NewRegistry(nil), nil, nil, nil, nil, nil, nil, nil)
	return service, registry
}

// TestTradeAccessSnapshotsAndForceCancels verifies bounded active-trade agency.
func TestTradeAccessSnapshotsAndForceCancels(t *testing.T) {
	service, registry := tradeAccessServiceForTest()
	session := &traderuntime.Session{
		RoomID: 9,
		First:  traderuntime.Participant{PlayerID: 1, UnitID: 11, Username: "one", Items: []int64{10}},
		Second: traderuntime.Participant{PlayerID: 2, UnitID: 22, Username: "two", Items: []int64{20}},
	}
	if !registry.Start(session) {
		t.Fatal("start fixture trade")
	}
	access := NewAccess(service, pluginruntime.NewScope("moderation"))
	snapshot, found := access.Active(1)
	if !found || snapshot.RoomID != 9 || snapshot.First.ItemIDs[0] != 10 {
		t.Fatalf("snapshot=%#v found=%v", snapshot, found)
	}
	snapshot.First.ItemIDs[0] = 99
	current, _ := session.Snapshot()
	if current.Items[0] != 10 {
		t.Fatal("snapshot mutation escaped into live trade")
	}
	if err := access.ForceCancel(1, "reported"); err != nil {
		t.Fatal(err)
	}
	if _, found := registry.Find(1); found {
		t.Fatal("trade remained active")
	}
}

// TestTradeAccessFailsClosedForDisabledScope verifies isolated scopes cannot act later.
func TestTradeAccessFailsClosedForDisabledScope(t *testing.T) {
	service, _ := tradeAccessServiceForTest()
	scope := pluginruntime.NewScope("disabled")
	scope.Disable()
	access := NewAccess(service, scope)
	if err := access.ForceCancel(1, "reported"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}
