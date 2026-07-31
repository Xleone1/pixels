package avatar

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// freezeKey identifies one unit's scheduled freeze generation.
type freezeKey struct {
	// roomID identifies the active room.
	roomID int64
	// entityKey identifies the room unit.
	entityKey int64
}

// executeControl applies movement-control effects through the room owner.
func (service *Service) executeControl(ctx context.Context, active *roomlive.Room, operation effect.AvatarOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	entityKey := event.ActorID
	if entityKey == 0 {
		entityKey = event.PlayerID
	}
	unit, found := active.UnitMotion(entityKey)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	switch operation {
	case effect.FreezeAvatar:
		frozen, err := active.SetUnitControl(entityKey, worldunit.ControlFrozen)
		if err != nil {
			return effect.Result{Status: effect.Skipped}, nil
		}
		duration := time.Duration(parameterAt(node, 0)) * time.Second
		if duration > 0 {
			token := service.nextFreeze(active.ID(), entityKey)
			active.Schedule(duration, func(time.Time) {
				if service.releaseFreeze(active.ID(), entityKey, token) {
					_, _ = active.ReleaseUnitControl(entityKey)
				}
			})
		}
		return resultApplied(), broadcast.RoomUnitStatuses(ctx, service.connections, active, []roomlive.UnitSnapshot{frozen}, 0)
	case effect.UnfreezeAvatar:
		service.invalidateFreeze(active.ID(), entityKey)
		released, err := active.ReleaseUnitControl(entityKey)
		if err != nil {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return resultApplied(), broadcast.RoomUnitStatuses(ctx, service.connections, active, []roomlive.UnitSnapshot{released}, 0)
	case effect.MoveRotateAvatar:
		point, valid := grid.PointInFront(unit.Position.Point, uint8(parameterAt(node, 0)))
		if !valid {
			return effect.Result{Status: effect.Blocked}, nil
		}
		rotation := worldunit.Rotation(parameterAt(node, 1))
		moved, err := active.TeleportUnit(entityKey, point, rotation, false)
		if err != nil {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return resultApplied(), broadcast.RoomUnitStatuses(ctx, service.connections, active, []roomlive.UnitSnapshot{moved}, 0)
	default:
		return effect.Result{Status: effect.Blocked}, nil
	}
}

// CleanupRoom releases freeze generations owned by a closed room.
func (service *Service) CleanupRoom(roomID int64) {
	service.freezeMutex.Lock()
	for key := range service.freezeTokens {
		if key.roomID == roomID {
			delete(service.freezeTokens, key)
		}
	}
	service.freezeMutex.Unlock()
}

// nextFreeze invalidates older timers and returns the current generation.
func (service *Service) nextFreeze(roomID int64, entityKey int64) uint64 {
	key := freezeKey{roomID: roomID, entityKey: entityKey}
	service.freezeMutex.Lock()
	service.freezeTokens[key]++
	token := service.freezeTokens[key]
	service.freezeMutex.Unlock()
	return token
}

// invalidateFreeze prevents scheduled releases without retaining a tombstone.
func (service *Service) invalidateFreeze(roomID int64, entityKey int64) {
	service.freezeMutex.Lock()
	delete(service.freezeTokens, freezeKey{roomID: roomID, entityKey: entityKey})
	service.freezeMutex.Unlock()
}

// releaseFreeze claims one timer only when it remains the newest generation.
func (service *Service) releaseFreeze(roomID int64, entityKey int64, token uint64) bool {
	key := freezeKey{roomID: roomID, entityKey: entityKey}
	service.freezeMutex.Lock()
	current := service.freezeTokens[key]
	if current == token {
		delete(service.freezeTokens, key)
	}
	service.freezeMutex.Unlock()
	return current == token
}

// parameterAt returns one configured integer or zero.
func parameterAt(node *configuration.Node, index int) int32 {
	if index < 0 || index >= len(node.Parameters.Values) {
		return 0
	}
	return node.Parameters.Values[index]
}
