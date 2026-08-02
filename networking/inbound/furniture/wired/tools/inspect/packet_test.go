package inspect

import (
	"math"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies entity inspection decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(4), codec.String("furniture"), codec.String("9223372036854775807"))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.EntityType != "furniture" || payload.EntityID != math.MaxInt64 {
		t.Fatalf("decode inspection: %#v %v", payload, err)
	}
}
