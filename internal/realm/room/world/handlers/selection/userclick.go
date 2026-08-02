// Package selection adapts room entity selection packets into commands.
package selection

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	userclickcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/userclick"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inuserclick "github.com/niflaot/pixels/networking/inbound/room/entities/userclick"
	"go.uber.org/zap"
)

// NewUserClick creates an avatar click packet handler.
func NewUserClick(handler userclickcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inuserclick.Decode(packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[userclickcmd.Command]{
			Command:  userclickcmd.Command{Handler: connection, RoomUnitID: int64(payload.RoomUnitID)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// RegisterUserClick adds the avatar click handler to a registry.
func RegisterUserClick(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inuserclick.Header, handler)
}
