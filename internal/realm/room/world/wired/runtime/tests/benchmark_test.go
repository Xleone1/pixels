package tests

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// BenchmarkNoCandidateDispatch measures the mandatory empty-event fast path.
func BenchmarkNoCandidateDispatch(benchmark *testing.B) {
	engine := newEngineForBenchmark(benchmark, []record.Config{{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1}}, &avatar{})
	event := trigger.Event{Kind: trigger.StateChanged, RoomID: 1}
	now := time.Now()
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		_, _ = engine.Process(context.Background(), event, now)
	}
}

// BenchmarkIndexedDispatch measures one indexed trigger and effect stack.
func BenchmarkIndexedDispatch(benchmark *testing.B) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "hello"},
	}
	service := &avatar{items: make([]int64, 0, 1)}
	engine := newEngineForBenchmark(benchmark, records, service)
	event := trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 4}
	now := time.Now()
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		service.items = service.items[:0]
		service.actors = service.actors[:0]
		_, _ = engine.Process(context.Background(), event, now)
	}
}

// BenchmarkSelectorDispatch measures one area selector fanning out to two actors.
func BenchmarkSelectorDispatch(benchmark *testing.B) {
	records := []record.Config{
		{
			ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room",
			X: 1, Y: 1,
		},
		{
			ItemID: 2, RoomID: 1, Interaction: "wf_slc_users_area",
			X: 1, Y: 1, IntParams: []int32{0, 0, 10, 10},
		},
		{
			ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message",
			X: 1, Y: 1, StringParam: "selected",
		},
	}
	service := &avatar{
		items: make([]int64, 0, 2), actors: make([]int64, 0, 2),
	}
	engine := wiredEngine(
		benchmark, records, effect.Services{Avatar: service},
	).WithDynamicDependencies(selection.New(100), nil)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		benchmark.Fatal(err)
	}
	event := trigger.Event{
		Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer,
		ActorID: 7, PlayerID: 7,
	}
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		service.items = service.items[:0]
		service.actors = service.actors[:0]
		if _, err := engine.Process(context.Background(), event, now); err != nil {
			benchmark.Fatal(err)
		}
	}
}

// newEngineForBenchmark adapts the test helper to benchmarks.
func newEngineForBenchmark(benchmark *testing.B, records []record.Config, service *avatar) *wiredruntime.Engine {
	benchmark.Helper()
	engine := newEngine(benchmark, records, service)
	if err := engine.Reload(context.Background(), 1, time.Now()); err != nil {
		benchmark.Fatal(err)
	}
	return engine
}
