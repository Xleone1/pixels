package open

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies panel open packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(88, "variables")
	if err != nil {
		t.Fatalf("encode open: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 88 || values[1].String != "variables" {
		t.Fatalf("decode open: %#v %v", values, err)
	}
}
