package event

import (
	"context"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// TestLifecycleAndCommandDispatchersForwardTypedSnapshots verifies non-cancellable facts.
func TestLifecycleAndCommandDispatchersForwardTypedSnapshots(t *testing.T) {
	local := bus.New()
	hub := NewHub(time.Second, zap.NewNop())
	hub.SetPlayerFinder(playerFinder{found: true})
	scope := pluginruntime.NewScope("observer")
	commands := 0
	currencies := 0
	_ = hub.listen(scope, sdkevent.CommandAttemptName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.CommandAttempt)
		if event.Input == "about" && event.Root == "about" {
			commands++
		}
		return nil
	})
	_ = hub.listen(scope, sdkevent.CurrencyChangedName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.CurrencyChanged)
		if event.Player.ID == 7 && event.Amount == 20 && event.Delta == 5 {
			currencies++
		}
		return nil
	})
	hub.DispatchCommandAttempt(context.Background(), testPlayer(), "about", "about")
	if err := hub.RegisterCurrencyChanged(local, playerFinder{found: true}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{
		Name: currencychanged.Name,
		Payload: currencychanged.Payload{
			PlayerID: 7, CurrencyType: -1, Amount: 20, Delta: 5, ActorKind: "plugin",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if commands != 1 || currencies != 1 {
		t.Fatalf("commands=%d currencies=%d", commands, currencies)
	}
}

// TestCurrencyGrantResolvesTheCompletePlayerSnapshot verifies pre-persistence events include live identity fields.
func TestCurrencyGrantResolvesTheCompletePlayerSnapshot(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	hub.SetPlayerFinder(playerFinder{found: true})
	called := false
	_ = hub.listen(pluginruntime.NewScope("observer"), sdkevent.CurrencyGrantName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.CurrencyGrant)
		called = event.Player == testPlayer()
		return nil
	})

	amount, cancelled := hub.DispatchCurrencyGrant(context.Background(), sdkplugin.Player{ID: 7}, 5, 10, "plugin")
	if !called || cancelled || amount != 10 {
		t.Fatalf("called=%v cancelled=%v amount=%d", called, cancelled, amount)
	}
}

// TestLifecycleForwardersIgnoreInvalidOrUnavailablePlayers verifies safe bridge fallbacks.
func TestLifecycleForwardersIgnoreInvalidOrUnavailablePlayers(t *testing.T) {
	local := bus.New()
	hub := NewHub(time.Second, nil)
	if err := hub.RegisterCurrencyChanged(local, playerFinder{found: false}); err != nil {
		t.Fatal(err)
	}
	if err := hub.RegisterPlayerConnected(local, playerFinder{found: false}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: currencychanged.Name, Payload: "invalid"}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{
		Name:    currencychanged.Name,
		Payload: currencychanged.Payload{PlayerID: 99, CurrencyType: 5},
	}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{
		Name:    playerconnected.Name,
		Payload: playerconnected.Payload{PlayerID: 99},
	}); err != nil {
		t.Fatal(err)
	}
	if player := hub.player(99); player.ID != 99 {
		t.Fatalf("fallback player=%#v", player)
	}
}
