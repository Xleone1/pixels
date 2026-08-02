package variable

import (
	"context"
	"sync"
	"testing"
	"time"
)

// memoryStore persists variable test values in one deterministic slice.
type memoryStore struct {
	// values stores durable assignments.
	values []Value
}

// systemFixture provides one immutable room variable.
type systemFixture struct{}

// ResolveSystem resolves the fixture system variable.
func (systemFixture) ResolveSystem(roomID int64, scope Scope, scopeID int64, name string) (Value, bool) {
	if roomID == 1 && scope == ScopeRoom && scopeID == 1 && name == "@id" {
		return Value{RoomID: roomID, Scope: scope, ScopeID: scopeID, Name: name, IntValue: 1}, true
	}
	return Value{}, false
}

// ListSystem lists the fixture system variable.
func (fixture systemFixture) ListSystem(roomID int64, scope Scope, scopeID int64) []Value {
	value, found := fixture.ResolveSystem(roomID, scope, scopeID, "@id")
	if !found {
		return nil
	}
	return []Value{value}
}

// LoadRoom returns assignments for one room.
func (store *memoryStore) LoadRoom(_ context.Context, roomID int64) ([]Value, error) {
	result := make([]Value, 0)
	for _, value := range store.values {
		if value.RoomID == roomID {
			result = append(result, value)
		}
	}
	return result, nil
}

// Find returns one exact durable assignment.
func (store *memoryStore) Find(
	_ context.Context,
	roomID int64,
	scope Scope,
	scopeID int64,
	name string,
) (Value, bool, error) {
	for _, value := range store.values {
		if value.RoomID == roomID && value.Scope == scope &&
			value.ScopeID == scopeID && value.Name == name {
			return value, true, nil
		}
	}
	return Value{}, false, nil
}

// Set creates or replaces one assignment.
func (store *memoryStore) Set(_ context.Context, value Value) (Value, error) {
	now := time.Unix(100, 0)
	for index, current := range store.values {
		if sameVariable(current, value) {
			value.CreatedAt = current.CreatedAt
			value.UpdatedAt = now
			store.values[index] = value
			return value, nil
		}
	}
	value.CreatedAt, value.UpdatedAt = now, now
	store.values = append(store.values, value)
	return value, nil
}

// Delete removes one exact assignment.
func (store *memoryStore) Delete(
	_ context.Context,
	roomID int64,
	scope Scope,
	scopeID int64,
	name string,
) (bool, error) {
	for index, value := range store.values {
		if value.RoomID == roomID && value.Scope == scope &&
			value.ScopeID == scopeID && value.Name == name {
			store.values = append(store.values[:index], store.values[index+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// sameVariable reports whether two assignments share an immutable key.
func sameVariable(left Value, right Value) bool {
	return left.RoomID == right.RoomID && left.Scope == right.Scope &&
		left.ScopeID == right.ScopeID && left.Name == right.Name
}

// TestServiceLifecycle verifies warming, normalization, mutation, and cleanup.
func TestServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := &memoryStore{values: []Value{{
		RoomID: 1, Scope: ScopeUser, ScopeID: 7, Name: "SCORE", IntValue: 2,
	}}}
	service := New(store, 2)
	if err := service.LoadRoom(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if value, found := service.Get(1, ScopeUser, 7, "score"); !found || value.IntValue != 2 {
		t.Fatalf("warmed value=%+v found=%t", value, found)
	}
	changed, previous, err := service.Change(ctx, Value{
		RoomID: 1, Scope: ScopeUser, ScopeID: 7, Name: " SCORE ",
	}, 3)
	if err != nil || previous != 2 || changed.IntValue != 5 {
		t.Fatalf("change value=%+v previous=%d error=%v", changed, previous, err)
	}
	if _, err = service.Set(ctx, Value{
		RoomID: 1, Scope: ScopeRoom, ScopeID: 1, Name: "round", IntValue: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Set(ctx, Value{
		RoomID: 1, Scope: ScopeUser, ScopeID: 8, Name: "score",
	}); err != ErrLimit {
		t.Fatalf("limit error=%v", err)
	}
	previousValue, deleted, err := service.Delete(ctx, 1, ScopeUser, 7, "SCORE")
	if err != nil || !deleted || previousValue.IntValue != 5 ||
		service.Exists(1, ScopeUser, 7, "score") {
		t.Fatalf("delete previous=%+v deleted=%t error=%v", previousValue, deleted, err)
	}
	service.Close(1)
	if service.Exists(1, ScopeRoom, 1, "round") {
		t.Fatal("closed room retained its warmed assignment")
	}
}

// TestConcurrentChangesRemainAtomic verifies serialized room mutations.
func TestConcurrentChangesRemainAtomic(t *testing.T) {
	ctx := context.Background()
	service := New(&memoryStore{}, 10)
	if err := service.LoadRoom(ctx, 1); err != nil {
		t.Fatal(err)
	}
	const workers = 64
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			if _, _, err := service.Change(ctx, Value{
				RoomID: 1, Scope: ScopeRoom, ScopeID: 1, Name: "counter",
			}, 1); err != nil {
				t.Errorf("change: %v", err)
			}
		}()
	}
	group.Wait()
	value, found := service.Get(1, ScopeRoom, 1, "counter")
	if !found || value.IntValue != workers {
		t.Fatalf("counter=%+v found=%t", value, found)
	}
}

// TestColdDurableReferenceSupportsReadAndChange verifies cross-room access.
func TestColdDurableReferenceSupportsReadAndChange(t *testing.T) {
	ctx := context.Background()
	store := &memoryStore{values: []Value{{
		RoomID: 2, Scope: ScopeRoom, ScopeID: 2,
		Name: "score", IntValue: 4,
	}}}
	service := New(store, 10)
	value, found, err := service.GetDurable(
		ctx, 2, ScopeRoom, 2, "SCORE",
	)
	if err != nil || !found || value.IntValue != 4 {
		t.Fatalf("value=%+v found=%t err=%v", value, found, err)
	}
	changed, previous, err := service.Change(ctx, Value{
		RoomID: 2, Scope: ScopeRoom, ScopeID: 2, Name: "score",
	}, 3)
	if err != nil || previous != 4 || changed.IntValue != 7 {
		t.Fatalf(
			"changed=%+v previous=%d err=%v",
			changed, previous, err,
		)
	}
	if _, warmed := service.Get(2, ScopeRoom, 2, "score"); warmed {
		t.Fatal("cold reference populated an unowned room cache")
	}
}

// TestSystemVariablesAreInspectableAndReadOnly verifies live projections never reach persistence.
func TestSystemVariablesAreInspectableAndReadOnly(t *testing.T) {
	service := New(&memoryStore{}, 10)
	service.SetSystemProvider(systemFixture{})
	if value, found := service.Get(1, ScopeRoom, 1, "@id"); !found || value.IntValue != 1 {
		t.Fatalf("system value=%+v found=%t", value, found)
	}
	if values := service.List(1, ScopeRoom, 1); len(values) != 1 || values[0].Name != "@id" {
		t.Fatalf("system list=%+v", values)
	}
	if _, err := service.Set(context.Background(), Value{RoomID: 1, Scope: ScopeRoom, ScopeID: 1, Name: "@id"}); err != ErrReadOnly {
		t.Fatalf("set read-only error=%v", err)
	}
	if _, _, err := service.Change(context.Background(), Value{RoomID: 1, Scope: ScopeRoom, ScopeID: 1, Name: "@id"}, 1); err != ErrReadOnly {
		t.Fatalf("change read-only error=%v", err)
	}
	if _, _, err := service.Delete(context.Background(), 1, ScopeRoom, 1, "@id"); err != ErrReadOnly {
		t.Fatalf("delete read-only error=%v", err)
	}
}

// BenchmarkWarmedGet measures the allocation-free variable read path.
func BenchmarkWarmedGet(b *testing.B) {
	service := New(&memoryStore{}, 10)
	service.rooms[1] = map[key]Value{
		{scope: ScopeUser, scopeID: 7, name: "score"}: {
			RoomID: 1, Scope: ScopeUser, ScopeID: 7, Name: "score", IntValue: 9,
		},
	}
	b.ReportAllocs()
	for range b.N {
		value, found := service.Get(1, ScopeUser, 7, "score")
		if !found || value.IntValue != 9 {
			b.Fatal("missing warmed value")
		}
	}
}
