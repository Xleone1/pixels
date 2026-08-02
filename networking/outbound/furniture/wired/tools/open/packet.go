// Package open encodes requests to open WIRED Creator Tools.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools open identifier.
const Header uint16 = 6303

// Definition describes one panel open request.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("initialTab", codec.StringField)}

// Encode creates one Creator Tools open packet.
func Encode(roomID int64, initialTab string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomID)), codec.String(initialTab))
}
