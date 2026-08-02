// Package result encodes WIRED Creator Tools mutation results.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED Creator Tools mutation result identifier.
const Header uint16 = 6301

// Definition describes one mutation result.
var Definition = codec.Definition{codec.Named("success", codec.BooleanField), codec.Named("code", codec.StringField), codec.Named("document", codec.StringField)}

// Encode creates one Creator Tools mutation result packet.
func Encode(success bool, code string, document string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(success), codec.String(code), codec.String(document))
}
