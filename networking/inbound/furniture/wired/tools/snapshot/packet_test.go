package snapshot

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies strict snapshot request decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(42))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 42 {
		t.Fatalf("decode snapshot: %#v %v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
