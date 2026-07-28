package event

import (
	"context"
	"errors"

	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

// DispatchChat sends one cancellable pre-delivery chat event.
func (hub *Hub) DispatchChat(ctx context.Context, player sdkplugin.Player, roomID int64, text string) (string, bool) {
	event := sdkevent.NewChatSend(player, roomID, text)
	err := hub.Dispatch(ctx, event)
	return event.Text, errors.Is(err, ErrEventCancelled)
}

// DispatchCommandAttempt publishes one non-cancellable command detector event.
func (hub *Hub) DispatchCommandAttempt(ctx context.Context, player sdkplugin.Player, input string, root string) {
	if !hub.HasListeners(sdkevent.CommandAttemptName) {
		return
	}
	_ = hub.Dispatch(ctx, &sdkevent.CommandAttempt{Player: player, Input: input, Root: root})
}

// DispatchCurrencyGrant sends one cancellable signed balance mutation event.
func (hub *Hub) DispatchCurrencyGrant(ctx context.Context, player sdkplugin.Player, currencyType int32, amount int64, actorKind string) (int64, bool) {
	if player.ID > 0 && player.Username == "" {
		player = hub.player(player.ID)
	}
	event := sdkevent.NewCurrencyGrant(player, currencyType, amount, actorKind)
	err := hub.Dispatch(ctx, event)
	return event.Amount, errors.Is(err, ErrEventCancelled)
}

// RegisterCurrencyChanged forwards committed currency mutations.
func (hub *Hub) RegisterCurrencyChanged(subscriber bus.Subscriber, players PlayerFinder) error {
	_, err := subscriber.Subscribe(currencychanged.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(currencychanged.Payload)
		if !ok {
			return nil
		}
		player, found := players.Find(payload.PlayerID)
		if !found {
			player = sdkplugin.Player{ID: payload.PlayerID}
		}
		return hub.Dispatch(ctx, &sdkevent.CurrencyChanged{
			Player: player, CurrencyType: payload.CurrencyType, Amount: payload.Amount,
			Delta: payload.Delta, ActorKind: payload.ActorKind,
		})
	})
	return err
}

// RegisterPlayerConnected forwards the post-authentication internal notification.
func (hub *Hub) RegisterPlayerConnected(subscriber bus.Subscriber, players PlayerFinder) error {
	_, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok {
			return nil
		}
		player, found := players.Find(payload.PlayerID)
		if !found {
			return nil
		}
		return hub.Dispatch(ctx, &sdkevent.PlayerConnected{Player: player})
	})
	return err
}
