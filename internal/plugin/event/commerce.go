package event

import (
	"context"
	"errors"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	sdkevent "github.com/niflaot/pixels/sdk/event"
)

// DispatchFurniturePlace sends one cancellable furniture placement.
func (hub *Hub) DispatchFurniturePlace(ctx context.Context, params furnitureservice.PlaceParams) (furnitureservice.PlaceParams, bool) {
	event := &sdkevent.FurniturePlace{
		Player: hub.player(params.ActorPlayerID), ItemID: params.ItemID, RoomID: params.RoomID,
		X: params.Placement.X, Y: params.Placement.Y, Z: params.Placement.Z,
		Rotation: int(params.Placement.Rotation), WallPosition: params.WallPosition,
	}
	err := hub.Dispatch(ctx, event)
	params.Placement = furnituremodel.Placement{
		X: event.X, Y: event.Y, Z: event.Z, Rotation: furnituremodel.Rotation(event.Rotation),
	}
	params.WallPosition = event.WallPosition
	return params, errors.Is(err, ErrEventCancelled)
}

// DispatchCatalogPurchase sends one cancellable catalog purchase.
func (hub *Hub) DispatchCatalogPurchase(ctx context.Context, playerID int64, item catalogmodel.Item, amount int32) (int64, int64, int32, bool) {
	event := &sdkevent.CatalogPurchase{
		Player: hub.player(playerID), CatalogItemID: item.ID, Amount: amount,
		CreditsCost: item.CostCredits, PointsCost: item.CostPoints, PointsType: item.PointsType,
	}
	err := hub.Dispatch(ctx, event)
	return event.CreditsCost, event.PointsCost, event.PointsType, errors.Is(err, ErrEventCancelled)
}
