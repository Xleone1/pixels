package result

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies mutation result packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(true, "saved", `{}`)
	if err != nil {
		t.Fatalf("encode result: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || !values[0].Boolean || values[1].String != "saved" {
		t.Fatalf("decode result: %#v %v", values, err)
	}
}
