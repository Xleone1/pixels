// Package tests verifies complete canonical effect dispatch.
package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// services records each focused dispatch.
type services struct {
	// calls stores dispatch count.
	calls int
}

// variableStore provides deterministic variable persistence to dispatch tests.
type variableStore struct {
	// values stores assignments by their stable test key.
	values map[string]variable.Value
}

// LoadRoom returns no preexisting assignments.
func (store *variableStore) LoadRoom(context.Context, int64) ([]variable.Value, error) {
	return nil, nil
}

// Set stores one assignment with deterministic timestamps.
func (store *variableStore) Set(_ context.Context, value variable.Value) (variable.Value, error) {
	now := time.Unix(1, 0)
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	value.UpdatedAt = now
	store.values[variableKey(value.RoomID, value.Scope, value.ScopeID, value.Name)] = value
	return value, nil
}

// Delete removes one assignment.
func (store *variableStore) Delete(_ context.Context, roomID int64, scope variable.Scope, scopeID int64, name string) (bool, error) {
	key := variableKey(roomID, scope, scopeID, name)
	_, found := store.values[key]
	delete(store.values, key)
	return found, nil
}

// variableKey creates a collision-free key for the dispatch fixture.
func variableKey(roomID int64, scope variable.Scope, scopeID int64, name string) string {
	return fmt.Sprintf("%d:%d:%d:%s", roomID, scope, scopeID, name)
}

// ExecuteFurniture records furniture dispatch.
func (service *services) ExecuteFurniture(context.Context, effect.FurnitureOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteAvatar records avatar dispatch.
func (service *services) ExecuteAvatar(context.Context, effect.AvatarOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteBot records bot dispatch.
func (service *services) ExecuteBot(context.Context, effect.BotOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// ExecuteGame records game dispatch.
func (service *services) ExecuteGame(context.Context, effect.GameOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// Claim records reward dispatch.
func (service *services) Claim(context.Context, *configuration.Node, trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// TestAllCanonicalEffectsDispatch verifies no canonical effect is an unhandled no-op.
func TestAllCanonicalEffectsDispatch(t *testing.T) {
	service := &services{}
	store := &variableStore{values: make(map[string]variable.Value)}
	variables := variable.New(store, 100)
	if err := variables.LoadRoom(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	executor := effect.New(effect.Services{
		Furniture: service, Avatar: service, Bot: service, Game: service,
		Reward: service, Variables: variables,
	})
	count := 0
	for _, descriptor := range registry.CanonicalManifest() {
		if descriptor.Family != registry.FamilyEffect {
			continue
		}
		count++
		node := &configuration.Node{
			Descriptor: descriptor, RoomID: 1, Targets: []record.Target{{ItemID: 9}},
			Parameters: configuration.Parameters{Text: "dispatch", Values: []int32{1, 1}},
		}
		result, err := executor.Execute(
			context.Background(),
			node,
			trigger.Event{RoomID: 1, ActorID: 7, PlayerID: 7},
		)
		if err != nil || result.Status != effect.Applied {
			t.Fatalf("effect %s result = %+v, %v", descriptor.Key, result, err)
		}
	}
	if count != 43 || service.calls != 37 {
		t.Fatalf("effects=%d service calls=%d", count, service.calls)
	}
}
