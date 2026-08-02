package inspection

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies inspection packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, `{"entity":{"type":"room"}}`)
	if err != nil {
		t.Fatalf("encode inspection: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 1 {
		t.Fatalf("decode inspection: %#v %v", values, err)
	}
}
