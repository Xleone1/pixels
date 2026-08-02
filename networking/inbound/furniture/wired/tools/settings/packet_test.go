package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Creator Tools settings decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(4), codec.Bool(true), codec.Bool(false))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || !payload.LiveUpdates || payload.HideBoxes {
		t.Fatalf("decode settings: %#v %v", payload, err)
	}
}
