package model

import "testing"

// TestRotationValid verifies supported floor rotation values.
func TestRotationValid(t *testing.T) {
	valid := []Rotation{RotationNorth, RotationEast, RotationSouth, RotationWest}
	for _, rotation := range valid {
		if !rotation.Valid() {
			t.Fatalf("expected rotation %d to be valid", rotation)
		}
	}

	invalid := []Rotation{1, 3, 5, 7, -1, 8}
	for _, rotation := range invalid {
		if rotation.Valid() {
			t.Fatalf("expected rotation %d to be invalid", rotation)
		}
	}
}

// TestRotationNormalize verifies diagonal and out-of-range rotations round to a cardinal value.
func TestRotationNormalize(t *testing.T) {
	cases := map[Rotation]Rotation{
		0: RotationNorth, 1: RotationEast, 2: RotationEast, 3: RotationSouth,
		4: RotationSouth, 5: RotationWest, 6: RotationWest, 7: RotationNorth,
		8: RotationNorth, -1: RotationNorth, 9: RotationEast,
	}
	for input, want := range cases {
		if got := input.Normalize(); got != want {
			t.Fatalf("Rotation(%d).Normalize() = %d, want %d", input, got, want)
		}
		if !input.Normalize().Valid() {
			t.Fatalf("Rotation(%d).Normalize() = %d is not a valid floor rotation", input, input.Normalize())
		}
	}
}
