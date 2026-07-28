package event

import (
	"context"
	"errors"

	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	sdkevent "github.com/niflaot/pixels/sdk/event"
)

// DispatchTradeStart sends one cancellable trade opening request.
func (hub *Hub) DispatchTradeStart(ctx context.Context, session *traderuntime.Session) bool {
	first, second := hub.tradeParticipants(session)
	event := &sdkevent.TradeStart{RoomID: session.RoomID, First: first, Second: second}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// DispatchTradeConfirm sends one cancellable final trade settlement request.
func (hub *Hub) DispatchTradeConfirm(ctx context.Context, session *traderuntime.Session) bool {
	first, second := hub.tradeParticipants(session)
	event := &sdkevent.TradeConfirm{RoomID: session.RoomID, First: first, Second: second}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// DispatchTradeCancel sends one cancellable active trade closure request.
func (hub *Hub) DispatchTradeCancel(ctx context.Context, playerID int64, session *traderuntime.Session, reason string) bool {
	first, second := hub.tradeParticipants(session)
	event := &sdkevent.TradeCancel{
		RoomID: session.RoomID, Actor: hub.player(playerID), First: first, Second: second, Reason: reason,
	}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// tradeParticipants converts one live trade into immutable SDK participants.
func (hub *Hub) tradeParticipants(session *traderuntime.Session) (sdkevent.TradeParticipant, sdkevent.TradeParticipant) {
	first, second := session.Snapshot()
	return hub.tradeParticipant(first), hub.tradeParticipant(second)
}

// tradeParticipant converts one internal participant into an SDK snapshot.
func (hub *Hub) tradeParticipant(participant traderuntime.Participant) sdkevent.TradeParticipant {
	return sdkevent.TradeParticipant{
		Player: hub.player(participant.PlayerID), UnitID: participant.UnitID,
		Items: append([]int64(nil), participant.Items...), Accepted: participant.Accepted, Confirmed: participant.Confirmed,
	}
}
