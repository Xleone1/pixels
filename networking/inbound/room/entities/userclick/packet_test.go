package userclick

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies avatar click decoding and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(17))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RoomUnitID != 17 {
		t.Fatalf("payload=%+v error=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("wrong header error=%v", err)
	}
}

// TestDecodeRejectsInvalidPayload verifies exact payload validation.
func TestDecodeRejectsInvalidPayload(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); err == nil {
		t.Fatal("expected payload error")
	}
}
