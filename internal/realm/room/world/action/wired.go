package action

import (
	"context"

	roomacted "github.com/niflaot/pixels/internal/realm/room/world/events/acted"
)

const (
	// WiredActionWave reports a wave expression.
	WiredActionWave int32 = iota + 1
	// WiredActionKiss reports a blown kiss expression.
	WiredActionKiss
	// WiredActionLaugh reports a laugh expression.
	WiredActionLaugh
	// WiredActionJump reports a jump expression.
	WiredActionJump
	// WiredActionRespect reports a respect expression.
	WiredActionRespect
	// WiredActionSit reports a sitting posture.
	WiredActionSit
	// WiredActionStand reports a standing posture.
	WiredActionStand
	// WiredActionLay reports a lying posture.
	WiredActionLay
	// WiredActionSign reports a held sign pulse.
	WiredActionSign
	// WiredActionDance reports a persistent dance.
	WiredActionDance
)

// ValidWiredAction reports whether a value belongs to the client contract.
func ValidWiredAction(value int32) bool {
	return value >= WiredActionWave && value <= WiredActionDance
}

// publishAction emits a normalized WIRED action event.
func (service *Service) publishAction(ctx context.Context, roomID int64, playerID int64, action int32, value int32) error {
	return service.publish(ctx, roomacted.Name, roomacted.Payload{
		RoomID: roomID, PlayerID: playerID, Action: action, Value: value,
	})
}

// expressionAction maps protocol expressions to WIRED action types.
func expressionAction(expressionID int32) int32 {
	switch expressionID {
	case 1:
		return WiredActionWave
	case 2:
		return WiredActionKiss
	case 3:
		return WiredActionLaugh
	case 6:
		return WiredActionJump
	case 7:
		return WiredActionRespect
	default:
		return 0
	}
}
