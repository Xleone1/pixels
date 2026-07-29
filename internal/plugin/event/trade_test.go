package event

import (
	"context"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	"go.uber.org/zap"
)

// TestTradeDispatchersCancelEveryLifecycleGate verifies trade snapshots and vetoes.
func TestTradeDispatchersCancelEveryLifecycleGate(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	hub.SetPlayerFinder(playerFinder{found: true})
	scope := pluginruntime.NewScope("trades")
	session := &traderuntime.Session{
		RoomID: 3,
		First:  traderuntime.Participant{PlayerID: 7, UnitID: 1, Items: []int64{10}, Accepted: true},
		Second: traderuntime.Participant{PlayerID: 8, UnitID: 2, Items: []int64{20}, Accepted: true},
	}
	for _, name := range []string{sdkevent.TradeStartName, sdkevent.TradeConfirmName, sdkevent.TradeCancelName} {
		_ = hub.listen(scope, name, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
			current.(sdkevent.Cancellable).SetCancelled(true)
			return nil
		})
	}
	if !hub.DispatchTradeStart(context.Background(), session) {
		t.Fatal("trade start was not cancelled")
	}
	if !hub.DispatchTradeConfirm(context.Background(), session) {
		t.Fatal("trade confirm was not cancelled")
	}
	if !hub.DispatchTradeCancel(context.Background(), 7, session, "manual") {
		t.Fatal("trade cancel was not cancelled")
	}
}

// BenchmarkTradeDispatchersWithoutListeners measures original event guard allocations.
func BenchmarkTradeDispatchersWithoutListeners(b *testing.B) {
	hub := NewHub(time.Second, zap.NewNop())
	ctx := context.Background()
	session := &traderuntime.Session{}
	b.Run("start", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = hub.DispatchTradeStart(ctx, session)
		}
	})
	b.Run("confirm", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = hub.DispatchTradeConfirm(ctx, session)
		}
	})
	b.Run("cancel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = hub.DispatchTradeCancel(ctx, 7, session, "manual")
		}
	})
}
