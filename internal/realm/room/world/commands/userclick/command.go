// Package userclick validates deliberate room avatar selections.
package userclick

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	avatarclicked "github.com/niflaot/pixels/internal/realm/room/world/events/avatarclicked"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies room avatar click commands.
	Name command.Name = "room.avatar.click"
)

var (
	// ErrPlayerNotInRoom reports a click without active room presence.
	ErrPlayerNotInRoom = errors.New("player not in room")
)

// Command contains one room avatar click.
type Command struct {
	// Handler stores source connection context.
	Handler netconn.Context
	// RoomUnitID identifies the selected room-local unit.
	RoomUnitID int64
}

// Handler validates and publishes avatar clicks.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated session bindings.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Events publishes validated room events.
	Events bus.Publisher
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle validates and publishes one avatar click.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return ErrPlayerNotInRoom
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	target, found := active.UnitByID(envelope.Command.RoomUnitID)
	if !found || target.Kind != worldunit.KindPlayer || target.PlayerID <= 0 || target.PlayerID == player.ID() {
		return nil
	}
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(ctx, bus.Event{Name: avatarclicked.Name, Payload: avatarclicked.Payload{
		RoomID: roomID, PlayerID: player.ID(), TargetPlayerID: target.PlayerID,
	}})
}
