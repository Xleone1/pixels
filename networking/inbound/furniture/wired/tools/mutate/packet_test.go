package mutate

import (
	"math"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Creator Tools mutation field order.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(8), codec.String("set"),
		codec.Int32(3), codec.String("9223372036854775807"), codec.String("score"), codec.String("9223372036854775807"), codec.String("final"))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 8 || payload.Scope != 3 || payload.ScopeID != math.MaxInt64 || payload.IntValue != "9223372036854775807" || payload.StringValue != "final" {
		t.Fatalf("decode mutation: %#v %v", payload, err)
	}
}
