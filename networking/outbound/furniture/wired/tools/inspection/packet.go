// Package inspection encodes WIRED Creator Tools entity inspections.
package inspection

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools inspection identifier.
const Header uint16 = 6302

// Definition describes one versioned JSON inspection.
var Definition = codec.Definition{codec.Named("schemaVersion", codec.Int32Field), codec.Named("document", codec.StringField)}

// Encode creates one Creator Tools inspection packet.
func Encode(schemaVersion int32, document string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(schemaVersion), codec.String(document))
}
