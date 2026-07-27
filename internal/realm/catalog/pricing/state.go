// Package pricing owns the process-local catalog price bypass.
package pricing

import (
	"sync/atomic"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// State stores the current process-local catalog pricing mode.
type State struct {
	// freeItems reports whether player-facing offers cost no currency.
	freeItems atomic.Bool
}

// New creates pricing state initialized from process configuration.
func New(freeItems bool) *State {
	state := &State{}
	state.freeItems.Store(freeItems)

	return state
}

// FreeItems reports whether every catalog offer is currently free.
func (state *State) FreeItems() bool {
	return state != nil && state.freeItems.Load()
}

// ToggleFreeItems atomically changes the current pricing mode.
func (state *State) ToggleFreeItems() bool {
	if state == nil {
		return false
	}
	for {
		current := state.freeItems.Load()
		if state.freeItems.CompareAndSwap(current, !current) {
			return !current
		}
	}
}

// Apply returns one offer projected under the current pricing mode.
func (state *State) Apply(item catalogmodel.Item) catalogmodel.Item {
	if state.FreeItems() {
		item.CostCredits = 0
		item.CostPoints = 0
	}

	return item
}
