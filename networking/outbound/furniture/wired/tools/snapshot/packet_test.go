package snapshot

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies snapshot packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, `{"room":{"id":7}}`)
	if err != nil {
		t.Fatalf("encode snapshot: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 1 || values[1].String == "" {
		t.Fatalf("decode snapshot: %#v %v", values, err)
	}
}
