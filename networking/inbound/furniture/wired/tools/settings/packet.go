// Package settings decodes WIRED Creator Tools settings requests.
package settings

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools settings identifier.
const Header uint16 = 6303

// Definition describes one settings request.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("liveUpdates", codec.BooleanField), codec.Named("hideBoxes", codec.BooleanField)}

// Payload stores one settings request.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// LiveUpdates enables client snapshot refreshes.
	LiveUpdates bool
	// HideBoxes requests hidden WIRED furniture.
	HideBoxes bool
}

// Decode decodes one settings request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: int64(values[0].Int32), LiveUpdates: values[1].Boolean, HideBoxes: values[2].Boolean}, nil
}
