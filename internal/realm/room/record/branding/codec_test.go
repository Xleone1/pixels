package branding

import (
	"strings"
	"testing"
)

// TestEncodeExtraDataUsesNitroKeys verifies the stable renderer map contract.
func TestEncodeExtraDataUsesNitroKeys(t *testing.T) {
	encoded := encodeExtraData(Mutation{State: 2, ImageURL: "https://cdn.example/banner.png", ClickURL: "https://example.com", OffsetX: -1, OffsetY: 2, OffsetZ: 3}, true)
	for _, fragment := range []string{`"state":"2"`, `"imageUrl":"https://cdn.example/banner.png"`, `"clickUrl":"https://example.com"`, `"offsetX":"-1"`, `"offsetY":"2"`, `"offsetZ":"3"`} {
		if !strings.Contains(encoded, fragment) {
			t.Fatalf("expected %q in %s", fragment, encoded)
		}
	}
}
