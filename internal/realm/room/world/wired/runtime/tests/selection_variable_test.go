package tests

import (
	"context"
	"testing"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// selectorProvider exposes deterministic actor selection without condition facts.
type selectorProvider struct{}

// View reports that no classic condition view is needed.
func (selectorProvider) View(int64) (condition.View, bool) { return nil, false }

// SelectionView returns the deterministic selector fixture.
func (selectorProvider) SelectionView(int64) (selection.View, bool) {
	return selectorFixture{}, true
}

// selectorFixture returns actors eight and nine for area queries.
type selectorFixture struct{}

// UsersByAction returns no actors.
func (selectorFixture) UsersByAction(int32) []int64 { return nil }

// UsersByName returns no actors.
func (selectorFixture) UsersByName(string) []int64 { return nil }

// UsersOnFurniture returns no actors.
func (selectorFixture) UsersOnFurniture([]record.Target) []int64 { return nil }

// UsersInGroup returns no actors.
func (selectorFixture) UsersInGroup() []int64 { return nil }

// UsersByHanditem returns no actors.
func (selectorFixture) UsersByHanditem(int32) []int64 { return nil }

// UsersByType returns no actors.
func (selectorFixture) UsersByType(int32) []int64 { return nil }

// UsersInArea returns deterministic selected actors.
func (selectorFixture) UsersInArea(int, int, int, int) []int64 {
	return []int64{9, 8}
}

// UsersByTeam returns no actors.
func (selectorFixture) UsersByTeam(int32) []int64 { return nil }

// FurniByType returns no furniture.
func (selectorFixture) FurniByType([]record.Target) []record.Target { return nil }

// FurniByAltitude returns no furniture.
func (selectorFixture) FurniByAltitude(int32, int32) []record.Target { return nil }

// FurniOnFurniture returns no furniture.
func (selectorFixture) FurniOnFurniture([]record.Target) []record.Target { return nil }

// FurniInArea returns no furniture.
func (selectorFixture) FurniInArea(int, int, int, int) []record.Target { return nil }

// FurniByVariable returns no furniture.
func (selectorFixture) FurniByVariable(string) []record.Target { return nil }

// UsersByVariable returns no actors.
func (selectorFixture) UsersByVariable(string) []int64 { return nil }

// variableStore persists one assignment for derived-event integration.
type variableStore struct {
	// value stores the current assignment.
	value variable.Value
	// found reports whether value exists.
	found bool
}

// LoadRoom returns the current assignment when present.
func (store *variableStore) LoadRoom(
	_ context.Context,
	roomID int64,
) ([]variable.Value, error) {
	if !store.found || store.value.RoomID != roomID {
		return nil, nil
	}
	return []variable.Value{store.value}, nil
}

// Find returns one exact durable assignment.
func (store *variableStore) Find(
	_ context.Context,
	roomID int64,
	scope variable.Scope,
	scopeID int64,
	name string,
) (variable.Value, bool, error) {
	value := store.value
	found := store.found && value.RoomID == roomID &&
		value.Scope == scope && value.ScopeID == scopeID &&
		value.Name == name
	return value, found, nil
}

// Set persists one assignment with timestamps.
func (store *variableStore) Set(_ context.Context, value variable.Value) (variable.Value, error) {
	now := time.Now()
	if store.found {
		value.CreatedAt = store.value.CreatedAt
	} else {
		value.CreatedAt = now
	}
	value.UpdatedAt = now
	store.value, store.found = value, true
	return value, nil
}

// Delete removes the current assignment.
func (store *variableStore) Delete(context.Context, int64, variable.Scope, int64, string) (bool, error) {
	found := store.found
	store.value, store.found = variable.Value{}, false
	return found, nil
}

// TestSelectorTargetsOverrideTriggerActor verifies selector-driven effect fan-out.
func TestSelectorTargetsOverrideTriggerActor(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_slc_users_area", X: 1, Y: 1, IntParams: []int32{0, 0, 10, 10}},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "selected"},
	}
	service := &avatar{}
	engine := wiredEngine(t, records, effect.Services{Avatar: service})
	engine.WithDynamicDependencies(selection.New(100), nil)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(
		context.Background(),
		trigger.Event{
			Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer,
			ActorID: 7, PlayerID: 7,
		},
		now,
	)
	if err != nil || trace.Effects != 2 || service.calls != 2 {
		t.Fatalf("trace=%+v calls=%d err=%v", trace, service.calls, err)
	}
}

// TestSignalSelectorReceivesDerivedTargets verifies exact queued event context.
func TestSignalSelectorReceivesDerivedTargets(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{
			ItemID: 2, RoomID: 1, Interaction: "wf_slc_users_area",
			X: 1, Y: 1, IntParams: []int32{0, 0, 10, 10},
		},
		{
			ItemID: 3, RoomID: 1, Interaction: "wf_act_send_signal",
			X: 1, Y: 1, StringParam: "selected",
		},
		{
			ItemID: 4, RoomID: 1, Interaction: "wf_trg_recv_signal",
			X: 2, Y: 2, StringParam: "selected",
		},
		{
			ItemID: 5, RoomID: 1, Interaction: "wf_slc_users_signal",
			X: 2, Y: 2,
		},
		{
			ItemID: 6, RoomID: 1, Interaction: "wf_act_show_message",
			X: 2, Y: 2, StringParam: "signal",
		},
	}
	service := &avatar{}
	engine := wiredEngine(
		t, records, effect.Services{Avatar: service},
	).WithDynamicDependencies(selection.New(100), nil)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(
		context.Background(),
		trigger.Event{
			Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer,
			ActorID: 7, PlayerID: 7,
		},
		now,
	)
	if err != nil || trace.Stacks != 2 || trace.Effects != 4 ||
		len(service.actors) != 2 ||
		service.actors[0] != 8 || service.actors[1] != 9 {
		t.Fatalf("trace=%+v actors=%v err=%v", trace, service.actors, err)
	}
}

// TestSignalAndVariableEventsRemainInsideTrace verifies both derived-event pipelines.
func TestSignalAndVariableEventsRemainInsideTrace(t *testing.T) {
	assignments := &variableStore{}
	variables := variable.New(assignments, 10)
	if err := variables.LoadRoom(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_send_signal", X: 1, Y: 1, StringParam: "next"},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_give_var", X: 1, Y: 1, StringParam: "score", IntParams: []int32{3, 5}},
		{ItemID: 4, RoomID: 1, Interaction: "wf_trg_recv_signal", X: 2, Y: 2, StringParam: "next"},
		{ItemID: 5, RoomID: 1, Interaction: "wf_act_show_message", X: 2, Y: 2, StringParam: "signal"},
		{ItemID: 6, RoomID: 1, Interaction: "wf_trg_var_changed", X: 3, Y: 3, StringParam: "score"},
		{ItemID: 7, RoomID: 1, Interaction: "wf_act_show_message", X: 3, Y: 3, StringParam: "variable"},
	}
	service := &avatar{}
	engine := wiredEngine(
		t, records, effect.Services{Avatar: service, Variables: variables},
	).WithDynamicDependencies(nil, variables)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(
		context.Background(),
		trigger.Event{
			Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer,
			ActorID: 7, PlayerID: 7,
		},
		now,
	)
	value, found := variables.Get(1, variable.ScopeRoom, 1, "score")
	if err != nil || trace.Stacks != 3 || trace.Effects != 4 ||
		service.calls != 2 || !found || value.IntValue != 5 {
		t.Fatalf(
			"trace=%+v calls=%d variable=%+v found=%v err=%v",
			trace, service.calls, value, found, err,
		)
	}
}

// TestReferenceVariableMutatesRemoteRoom verifies cold cross-room definitions.
func TestReferenceVariableMutatesRemoteRoom(t *testing.T) {
	assignments := &variableStore{
		value: variable.Value{
			RoomID: 2, Scope: variable.ScopeRoom, ScopeID: 2,
			Name: "score", IntValue: 4,
		},
		found: true,
	}
	variables := variable.New(assignments, 10)
	if err := variables.LoadRoom(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	records := []record.Config{
		{
			ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room",
			X: 1, Y: 1,
		},
		{
			ItemID: 2, RoomID: 1, Interaction: "wf_var_reference",
			X: 1, Y: 1, IntParams: []int32{2}, StringParam: "score",
		},
		{
			ItemID: 3, RoomID: 1, Interaction: "wf_act_change_var_val",
			X: 1, Y: 1, IntParams: []int32{4, 3}, StringParam: "alias",
		},
	}
	engine := wiredEngine(
		t, records, effect.Services{Variables: variables},
	).WithDynamicDependencies(nil, variables)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(
		context.Background(),
		trigger.Event{
			Kind: trigger.EnterRoom, RoomID: 1,
			ActorKind: trigger.ActorPlayer, ActorID: 7, PlayerID: 7,
		},
		now,
	)
	value, found, findErr := variables.GetDurable(
		context.Background(), 2, variable.ScopeRoom, 2, "score",
	)
	if err != nil || findErr != nil || trace.Effects != 1 ||
		!found || value.IntValue != 7 {
		t.Fatalf(
			"trace=%+v value=%+v found=%t err=%v findErr=%v",
			trace, value, found, err, findErr,
		)
	}
}

// wiredEngine creates an engine with selected WIRED service boundaries.
func wiredEngine(
	t *testing.T,
	records []record.Config,
	services effect.Services,
) *wiredruntime.Engine {
	t.Helper()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	return wiredruntime.New(
		roomwired.Config{Enabled: true}, store{records: records}, compiler,
		effect.New(services), selectorProvider{}, nil, nil,
	)
}
