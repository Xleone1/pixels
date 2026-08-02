// Package userclick decodes deliberate room avatar selections.
package userclick

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies the Pixels avatar click compatibility packet.
	Header uint16 = 6304
)

// Payload contains the selected room-local unit identifier.
type Payload struct {
	// RoomUnitID identifies the selected avatar in the active room.
	RoomUnitID int32
}

// Definition describes the avatar click payload.
var Definition = codec.Definition{codec.Named("roomUnitId", codec.Int32Field)}

// Decode unpacks one avatar click request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomUnitID: values[0].Int32}, nil
}
