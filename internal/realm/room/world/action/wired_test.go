package action

import "testing"

// TestValidWiredActionKeepsSemanticBounds verifies the shared client contract.
func TestValidWiredActionKeepsSemanticBounds(t *testing.T) {
	for value := int32(1); value <= 10; value++ {
		if !ValidWiredAction(value) {
			t.Fatalf("ValidWiredAction(%d) = false", value)
		}
	}
	for _, value := range []int32{-1, 0, 11, 12} {
		if ValidWiredAction(value) {
			t.Fatalf("ValidWiredAction(%d) = true", value)
		}
	}
}
