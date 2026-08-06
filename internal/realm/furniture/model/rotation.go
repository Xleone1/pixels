package model

// Rotation stores a floor furniture instance rotation.
type Rotation int16

const (
	// RotationNorth stores north-facing rotation.
	RotationNorth Rotation = 0

	// RotationEast stores east-facing rotation.
	RotationEast Rotation = 2

	// RotationSouth stores south-facing rotation.
	RotationSouth Rotation = 4

	// RotationWest stores west-facing rotation.
	RotationWest Rotation = 6
)

// Valid reports whether the rotation is one of the supported floor values.
func (rotation Rotation) Valid() bool {
	switch rotation {
	case RotationNorth, RotationEast, RotationSouth, RotationWest:
		return true
	default:
		return false
	}
}

// Normalize rounds an arbitrary eight-direction rotation to the nearest supported floor value.
// Nitro sometimes forwards a placed item's diagonal facing (odd values 1, 3, 5, or 7) instead of
// one of the four cardinal directions floor placement accepts, most visibly for furniture whose
// own widget drives placement (room branding anchors) rather than the standard rotate control.
func (rotation Rotation) Normalize() Rotation {
	value := int(rotation) % 8
	if value < 0 {
		value += 8
	}

	return Rotation((value + 1) / 2 * 2 % 8)
}
