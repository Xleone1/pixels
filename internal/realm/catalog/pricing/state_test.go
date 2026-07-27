package pricing

import (
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// TestStateTogglesAndProjectsPrices verifies the process-local override.
func TestStateTogglesAndProjectsPrices(t *testing.T) {
	state := New(false)
	item := catalogmodel.Item{CostCredits: 10, CostPoints: 4}
	if state.FreeItems() || state.Apply(item) != item {
		t.Fatal("paid mode changed the durable price")
	}
	if enabled := state.ToggleFreeItems(); !enabled {
		t.Fatal("expected free mode")
	}
	free := state.Apply(item)
	if free.CostCredits != 0 || free.CostPoints != 0 {
		t.Fatalf("free item=%#v", free)
	}
	if enabled := state.ToggleFreeItems(); enabled || state.Apply(item) != item {
		t.Fatal("expected restored paid mode")
	}
}

// TestNilStateRemainsPaid verifies zero-value service compatibility.
func TestNilStateRemainsPaid(t *testing.T) {
	var state *State
	item := catalogmodel.Item{CostCredits: 10, CostPoints: 4}
	if state.FreeItems() || state.ToggleFreeItems() || state.Apply(item) != item {
		t.Fatal("nil state must remain paid")
	}
}
