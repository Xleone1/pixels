// Package inspect decodes WIRED Creator Tools entity inspection requests.
package inspect

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the WIRED Creator Tools inspection request identifier.
const Header uint16 = 6302

// Definition describes one entity inspection request.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("entityType", codec.StringField), codec.Named("entityId", codec.StringField)}

// Payload stores one entity inspection request.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// EntityType identifies room, user, or furniture.
	EntityType string
	// EntityID identifies the target entity.
	EntityID int64
}

// Decode decodes one entity inspection request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	entityID, err := strconv.ParseInt(values[2].String, 10, 64)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: int64(values[0].Int32), EntityType: values[1].String, EntityID: entityID}, nil
}
