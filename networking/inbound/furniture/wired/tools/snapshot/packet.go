// Package snapshot decodes WIRED Creator Tools snapshot requests.
package snapshot

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools snapshot request identifier.
const Header uint16 = 6300

// Definition describes a snapshot request.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Payload stores one snapshot request.
type Payload struct {
	// RoomID identifies the requested room.
	RoomID int64
}

// Decode decodes one snapshot request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: int64(values[0].Int32)}, nil
}
