package selection

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// filterStore returns immutable variable-filter fixtures.
type filterStore struct {
	// values stores fixture assignments.
	values []variable.Value
}

// LoadRoom returns fixture assignments.
func (store filterStore) LoadRoom(context.Context, int64) ([]variable.Value, error) {
	return store.values, nil
}

// Set is unused by read-only filter tests.
func (filterStore) Set(
	context.Context,
	variable.Value,
) (variable.Value, error) {
	return variable.Value{}, nil
}

// Delete is unused by read-only filter tests.
func (filterStore) Delete(
	context.Context,
	int64,
	variable.Scope,
	int64,
	string,
) (bool, error) {
	return false, nil
}

// TestFilterVariablesOrdersAndLimitsActors verifies highest-value selection.
func TestFilterVariablesOrdersAndLimitsActors(t *testing.T) {
	now := time.Unix(100, 0)
	variables := variable.New(filterStore{values: []variable.Value{
		{
			RoomID: 1, Scope: variable.ScopeUser, ScopeID: 1,
			Name: "score", IntValue: 3, CreatedAt: now, UpdatedAt: now,
		},
		{
			RoomID: 1, Scope: variable.ScopeUser, ScopeID: 2,
			Name: "score", IntValue: 9, CreatedAt: now, UpdatedAt: now,
		},
	}}, 10)
	if err := variables.LoadRoom(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	extra := &configuration.Node{
		Descriptor: registry.Descriptor{Key: "wf_xtra_filter_users_by_var"},
		Parameters: configuration.Parameters{
			Values: []int32{0, 0, 1},
			Text:   "score",
		},
	}
	result := FilterVariables(
		&configuration.Stack{Extras: []*configuration.Node{extra}},
		Selection{Actors: []int64{1, 2}},
		trigger.Event{RoomID: 1, ActorID: 1},
		variables,
	)
	if len(result.Actors) != 1 || result.Actors[0] != 2 {
		t.Fatalf("filtered actors=%v", result.Actors)
	}
}
