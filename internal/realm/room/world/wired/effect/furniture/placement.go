package furniture

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// setAltitude persists an exact base height when no furniture is stacked above.
func (service *Service) setAltitude(ctx context.Context, active *roomlive.Room, itemID int64, node *configuration.Node, actorID int64) (bool, error) {
	item, found := active.FurnitureItem(itemID)
	if !found || !active.CanChangeFurnitureHeight(itemID) || len(node.Parameters.Values) == 0 {
		return false, nil
	}
	units := float64(node.Parameters.Values[0]) / 100
	if units < 0 || units > 100 {
		return false, nil
	}
	next := grid.HeightFromUnits(units)
	if item.Z == next {
		return false, nil
	}
	durable, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil || !found {
		return false, err
	}
	if actorID <= 0 {
		actorID = durable.OwnerPlayerID
	}
	_, err = service.furniture.Move(ctx, furnitureservice.MoveParams{
		ItemID: item.ID, ActorPlayerID: actorID, RoomID: active.ID(),
		Placement: furnituremodel.Placement{
			X: int(item.Point.X), Y: int(item.Point.Y), Z: next.Units(),
			Rotation: durableRotation(item.Rotation),
		},
	})
	if err != nil {
		return false, err
	}
	item.Z = next
	if _, err = active.ReloadFurniture(item.ID, &item); err != nil {
		return false, err
	}
	return true, service.broadcast(ctx, active, item)
}

// moveFurnitureToFurniture moves every source one tile toward the final target.
func (service *Service) moveFurnitureToFurniture(ctx context.Context, active *roomlive.Room, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	if len(node.Targets) < 2 {
		return effect.Result{Status: effect.Skipped}, nil
	}
	destination, found := active.FurnitureItem(node.Targets[len(node.Targets)-1].ItemID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	result := effect.Result{Status: effect.Skipped}
	for _, target := range node.Targets[:len(node.Targets)-1] {
		item, exists := active.FurnitureItem(target.ItemID)
		if !exists {
			continue
		}
		points := cardinal(item.Point)
		sortByDistance(points, destination.Point, false)
		for _, point := range points {
			mutation, err := service.placeMovement(ctx, active, item, point, item.Rotation, event)
			if err != nil {
				return effect.Result{Status: effect.Blocked}, err
			}
			if mutation.changed {
				result.Status = effect.Applied
				break
			}
		}
	}
	return result, nil
}
