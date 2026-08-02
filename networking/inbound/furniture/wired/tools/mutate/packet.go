// Package mutate decodes WIRED Creator Tools variable mutations.
package mutate

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

// Header is the WIRED Creator Tools mutation identifier.
const Header uint16 = 6301

// Definition describes one variable mutation.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field), codec.Named("operation", codec.StringField),
	codec.Named("scope", codec.Int32Field), codec.Named("scopeId", codec.StringField),
	codec.Named("name", codec.StringField), codec.Named("intValue", codec.StringField),
	codec.Named("stringValue", codec.StringField),
}

// Payload stores one variable mutation request.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// Operation identifies set, change, or delete.
	Operation string
	// Scope identifies the variable target family.
	Scope int32
	// ScopeID identifies the target.
	ScopeID int64
	// Name identifies the variable.
	Name string
	// IntValue stores the decimal integer value or delta.
	IntValue string
	// StringValue stores the textual value.
	StringValue string
}

// Decode decodes one variable mutation.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	scopeID, err := strconv.ParseInt(values[3].String, 10, 64)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: int64(values[0].Int32), Operation: values[1].String,
		Scope: values[2].Int32, ScopeID: scopeID, Name: values[4].String,
		IntValue: values[5].String, StringValue: values[6].String}, nil
}
