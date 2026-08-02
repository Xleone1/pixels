package projectile

import (
	"context"
	"errors"
	"time"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// New creates a bounded custom projectile service.
func New(config roomwired.Config, rooms *roomlive.Registry, furniture *furnitureservice.Service, connections *netconn.Registry) *Service {
	normalized := config.Normalize()
	return &Service{
		rooms: rooms, furniture: furniture, connections: connections,
		maximum: normalized.MaxProjectilesPerRoom, maximumDistance: normalized.MaxProjectileDistance,
		maximumDuration: normalized.MaxProjectileDurationPulses,
	}
}

// ExecuteExtras launches configured projectiles after stack selection and conditions pass.
func (service *Service) ExecuteExtras(_ context.Context, extras []*configuration.Node, event trigger.Event, resolved selection.Selection, _ time.Time) error {
	if service == nil || service.rooms == nil || event.RoomID <= 0 {
		return nil
	}
	active, found := service.rooms.Find(event.RoomID)
	if !found {
		return nil
	}
	riderAvailable := event.ActorID > 0
	for _, extra := range extras {
		if extra == nil || extra.Descriptor.Key != Key {
			continue
		}
		targets := extra.Targets
		if resolved.FurniResolved {
			targets = resolved.Furni
		}
		for _, target := range targets {
			riderID := int64(0)
			if riderAvailable {
				if _, present := active.Unit(event.ActorID); present {
					riderID = event.ActorID
				}
			}
			if service.launch(active, extra, target.ItemID, riderID) {
				riderAvailable = false
			}
		}
	}
	return nil
}

// Cycle advances active projectiles from the owning room cycle.
func (service *Service) Cycle(ctx context.Context, roomID int64, _ time.Time) error {
	if service == nil {
		return nil
	}
	value, found := service.states.Load(roomID)
	if !found {
		return nil
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return nil
	}
	state := value.(*roomState)
	state.mutex.Lock()
	defer state.mutex.Unlock()
	var result error
	for index := range state.flights {
		current := &state.flights[index]
		if !current.used {
			continue
		}
		finished, err := service.advance(ctx, active, current)
		result = errors.Join(result, err)
		if finished {
			service.remember(state, current.snapshot)
			*current = flight{}
			state.active--
		}
	}
	return result
}

// Close persists current final positions and releases one room state.
func (service *Service) Close(roomID int64) {
	if service == nil {
		return
	}
	value, found := service.states.LoadAndDelete(roomID)
	if !found {
		return
	}
	state := value.(*roomState)
	state.mutex.Lock()
	defer state.mutex.Unlock()
	for index := range state.flights {
		current := &state.flights[index]
		if current.used {
			_ = service.persist(context.Background(), current)
		}
	}
}

// ResolveSystem returns one retained projectile system variable.
func (service *Service) ResolveSystem(roomID int64, scope variable.Scope, scopeID int64, name string) (variable.Value, bool) {
	if service == nil || scope != variable.ScopeFurni || scopeID <= 0 {
		return variable.Value{}, false
	}
	snapshot, found := service.status(roomID, scopeID)
	if !found {
		return variable.Value{}, false
	}
	value, found := snapshotValue(roomID, snapshot, name)
	return value, found
}

// ListSystem returns every retained projectile system variable for one furniture item.
func (service *Service) ListSystem(roomID int64, scope variable.Scope, scopeID int64) []variable.Value {
	if scope != variable.ScopeFurni {
		return nil
	}
	snapshot, found := service.status(roomID, scopeID)
	if !found {
		return nil
	}
	result := make([]variable.Value, 0, len(systemNames))
	for _, name := range systemNames {
		if value, resolved := snapshotValue(roomID, snapshot, name); resolved {
			result = append(result, value)
		}
	}
	return result
}

// status returns one active or retained snapshot.
func (service *Service) status(roomID int64, itemID int64) (snapshot, bool) {
	value, found := service.states.Load(roomID)
	if !found {
		return snapshot{}, false
	}
	state := value.(*roomState)
	state.mutex.RLock()
	defer state.mutex.RUnlock()
	for index := range state.flights {
		if state.flights[index].used && state.flights[index].current.ID == itemID {
			return state.flights[index].snapshot, true
		}
	}
	for index := range state.history {
		if state.history[index].itemID == itemID {
			return state.history[index], true
		}
	}
	return snapshot{}, false
}
