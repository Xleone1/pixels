package projectile

import (
	"context"
	"errors"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// collision classifies a rejected swept step.
type collision uint8

const (
	// collisionNone reports a free swept step.
	collisionNone collision = iota
	// collisionUser reports another room unit in the destination footprint.
	collisionUser
	// collisionFurniture reports furniture or room geometry in the destination footprint.
	collisionFurniture
)

// advance advances one flight by at most one tile.
func (service *Service) advance(ctx context.Context, active *roomlive.Room, current *flight) (bool, error) {
	current.age++
	if current.age > current.maximumAge {
		return true, service.finish(ctx, active, current)
	}
	current.pulse++
	if current.pulse < current.pulsesPerStep {
		return false, nil
	}
	current.pulse = 0
	point, valid := grid.PointInFront(current.current.Point, current.direction)
	if !valid {
		current.snapshot.canMoveFreely = false
		return true, service.finish(ctx, active, current)
	}
	candidate := current.current
	candidate.Point = point
	height, blocked := service.destination(active, candidate, current.riderID)
	if blocked != collisionNone {
		current.snapshot.canMoveFreely = false
		if blocked == collisionUser {
			current.snapshot.userCollisions++
		} else {
			current.snapshot.furniCollisions++
		}
		return true, service.finish(ctx, active, current)
	}
	candidate.Z = height
	previous := current.current
	if _, err := active.ReloadFurniture(candidate.ID, &candidate); err != nil {
		current.snapshot.canMoveFreely = false
		return true, errors.Join(err, service.finish(ctx, active, current))
	}
	current.current = candidate
	current.remaining--
	current.snapshot.position = candidate.Point
	current.snapshot.altitude = candidate.Z
	current.snapshot.tilesTravelled++
	current.snapshot.canMoveFreely = current.remaining > 0
	projectionErr := service.project(ctx, active, previous, candidate)
	projectionErr = errors.Join(projectionErr, service.follow(ctx, active, current))
	if current.remaining == 0 {
		return true, errors.Join(projectionErr, service.finish(ctx, active, current))
	}
	return false, projectionErr
}

// destination resolves the swept footprint height and first collision class.
func (service *Service) destination(active *roomlive.Room, item worldfurniture.Item, riderID int64) (grid.Height, collision) {
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	maximum := grid.Height(0)
	for dy := 0; dy < length; dy++ {
		for dx := 0; dx < width; dx++ {
			point, valid := grid.NewPoint(int(item.Point.X)+dx, int(item.Point.Y)+dy)
			if !valid {
				return 0, collisionFurniture
			}
			column, err := active.SurfaceColumn(point)
			if err != nil {
				return 0, collisionFurniture
			}
			base, valid := projectileBase(column, item.ID)
			if !valid {
				return 0, collisionFurniture
			}
			if base > maximum {
				maximum = base
			}
			var ids [16]int64
			for _, itemID := range active.FurnitureIDsAt(point, ids[:0]) {
				if itemID != item.ID {
					return 0, collisionFurniture
				}
			}
		}
	}
	for _, unit := range active.Units() {
		if unit.EntityKey != riderID && footprintContains(item, unit.Position.Point) {
			return 0, collisionUser
		}
	}
	return maximum, collisionNone
}

// projectileBase resolves a base tile while rejecting foreign fixture sections.
func projectileBase(column surface.Column, itemID int64) (grid.Height, bool) {
	var height grid.Height
	found := false
	for index := 0; index < column.Len(); index++ {
		section, valid := column.Section(index)
		if !valid {
			continue
		}
		if section.Source() == surface.SourceBase {
			height, found = section.Z(), true
			continue
		}
		if section.SourceID() != itemID {
			return 0, false
		}
	}
	return height, found
}

// footprintContains reports whether one tile lies inside a rotated footprint.
func footprintContains(item worldfurniture.Item, point grid.Point) bool {
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	return int(point.X) >= int(item.Point.X) && int(point.X) < int(item.Point.X)+width &&
		int(point.Y) >= int(item.Point.Y) && int(point.Y) < int(item.Point.Y)+length
}

// follow projects the optional actor onto the current projectile top without reserving a slot.
func (service *Service) follow(ctx context.Context, active *roomlive.Room, current *flight) error {
	if current.riderID <= 0 {
		return nil
	}
	unit, err := active.RollUnit(current.riderID, worldpath.Position{Point: current.current.Point, Z: current.current.Top()})
	if err != nil {
		current.riderID = 0
		return nil
	}
	if service.connections == nil {
		return nil
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
}

// finish persists one final placement and rolls back the runtime projection on failure.
func (service *Service) finish(ctx context.Context, active *roomlive.Room, current *flight) error {
	current.snapshot.moving = false
	if err := service.persist(ctx, current); err != nil {
		rollbackErr := service.restoreRider(ctx, active, current)
		if _, reloadErr := active.ReloadFurniture(current.origin.ID, &current.origin); reloadErr != nil {
			rollbackErr = errors.Join(rollbackErr, reloadErr)
		} else {
			rollbackErr = errors.Join(rollbackErr, service.project(ctx, active, current.current, current.origin))
			current.current = current.origin
			current.snapshot.position = current.origin.Point
			current.snapshot.altitude = current.origin.Z
		}
		return errors.Join(err, rollbackErr)
	}
	return nil
}

// restoreRider returns the optional following actor to its pre-flight position.
func (service *Service) restoreRider(ctx context.Context, active *roomlive.Room, current *flight) error {
	if current.riderID <= 0 {
		return nil
	}
	unit, err := active.RollUnit(current.riderID, current.riderOrigin)
	if err != nil || service.connections == nil {
		return err
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
}

// persist stores only the final projectile placement.
func (service *Service) persist(ctx context.Context, current *flight) error {
	if service.furniture == nil || current.current.Point == current.origin.Point && current.current.Z == current.origin.Z {
		return nil
	}
	_, err := service.furniture.Move(ctx, furnitureservice.MoveParams{
		ItemID: current.current.ID, ActorPlayerID: current.ownerID, RoomID: current.roomID,
		Placement: furnituremodel.Placement{X: int(current.current.Point.X), Y: int(current.current.Point.Y), Z: current.current.Z.Units(), Rotation: furnituremodel.Rotation(current.current.Rotation)},
	})
	return err
}

// project broadcasts normal furniture and height-map packets for one transient step.
func (service *Service) project(ctx context.Context, active *roomlive.Room, previous worldfurniture.Item, current worldfurniture.Item) error {
	if service.connections == nil {
		return nil
	}
	packet, err := outupdate.Encode(outupdate.FloorItem{
		ID: current.ID, SpriteID: current.Definition.SpriteID, X: int(current.Point.X), Y: int(current.Point.Y),
		Rotation: int(current.Rotation), Z: current.Z.String(), ExtraHeight: current.Top().String(),
		ExtraData: current.ExtraData, OwnerID: current.OwnerPlayerID,
	})
	if err != nil {
		return err
	}
	err = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	points := append(worldfurniture.Footprint(previous.Point, previous.Definition.Width, previous.Definition.Length, previous.Rotation),
		worldfurniture.Footprint(current.Point, current.Definition.Width, current.Definition.Length, current.Rotation)...)
	return errors.Join(err, broadcast.RoomHeightMapUpdate(ctx, service.connections, active, points, 0))
}
