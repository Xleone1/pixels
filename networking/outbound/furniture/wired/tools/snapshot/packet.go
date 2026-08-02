// Package snapshot encodes WIRED Creator Tools snapshots.
package snapshot

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools snapshot identifier.
const Header uint16 = 6300

// Definition describes one versioned JSON snapshot.
var Definition = codec.Definition{codec.Named("schemaVersion", codec.Int32Field), codec.Named("document", codec.StringField)}

// Encode creates one Creator Tools snapshot packet.
func Encode(schemaVersion int32, document string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(schemaVersion), codec.String(document))
}
