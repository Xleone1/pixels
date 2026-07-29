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
	if !hub.HasListeners(sdkevent.FurniturePlaceName) {
		return params, false
	}
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

// DispatchFurnitureMove sends one cancellable furniture movement.
func (hub *Hub) DispatchFurnitureMove(ctx context.Context, params furnitureservice.MoveParams) (furnitureservice.MoveParams, bool) {
	if !hub.HasListeners(sdkevent.FurnitureMoveName) {
		return params, false
	}
	event := &sdkevent.FurnitureMove{
		Player: hub.player(params.ActorPlayerID), ItemID: params.ItemID, RoomID: params.RoomID,
		X: params.Placement.X, Y: params.Placement.Y, Z: params.Placement.Z,
		Rotation: int(params.Placement.Rotation), WallPosition: params.WallPosition,
	}
	err := hub.Dispatch(ctx, event)
	params.Placement = furnituremodel.Placement{X: event.X, Y: event.Y, Z: event.Z, Rotation: furnituremodel.Rotation(event.Rotation)}
	params.WallPosition = event.WallPosition
	return params, errors.Is(err, ErrEventCancelled)
}

// DispatchFurniturePickup sends one cancellable furniture pickup.
func (hub *Hub) DispatchFurniturePickup(ctx context.Context, params furnitureservice.PickupParams) bool {
	if !hub.HasListeners(sdkevent.FurniturePickupName) {
		return false
	}
	event := &sdkevent.FurniturePickup{Player: hub.player(params.ActorPlayerID), ItemID: params.ItemID, RoomID: params.RoomID}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// DispatchCatalogPurchase sends one cancellable catalog purchase.
func (hub *Hub) DispatchCatalogPurchase(ctx context.Context, playerID int64, item catalogmodel.Item, amount int32) (int64, int64, int32, bool) {
	if !hub.HasListeners(sdkevent.CatalogPurchaseName) {
		return item.CostCredits, item.CostPoints, item.PointsType, false
	}
	event := &sdkevent.CatalogPurchase{
		Player: hub.player(playerID), CatalogItemID: item.ID, Amount: amount,
		CreditsCost: item.CostCredits, PointsCost: item.CostPoints, PointsType: item.PointsType,
	}
	err := hub.Dispatch(ctx, event)
	return event.CreditsCost, event.PointsCost, event.PointsType, errors.Is(err, ErrEventCancelled)
}

// DispatchMarketplaceList sends one cancellable marketplace listing.
func (hub *Hub) DispatchMarketplaceList(ctx context.Context, playerID int64, itemID int64, rawPrice int64) (int64, bool) {
	if !hub.HasListeners(sdkevent.MarketplaceListName) {
		return rawPrice, false
	}
	event := &sdkevent.MarketplaceList{Player: hub.player(playerID), ItemID: itemID, RawPrice: rawPrice}
	err := hub.Dispatch(ctx, event)
	return event.RawPrice, errors.Is(err, ErrEventCancelled)
}

// DispatchMarketplaceBuy sends one cancellable marketplace purchase.
func (hub *Hub) DispatchMarketplaceBuy(ctx context.Context, buyerID int64, listingID int64, sellerID int64, buyerPrice int64) (int64, bool) {
	if !hub.HasListeners(sdkevent.MarketplaceBuyName) {
		return buyerPrice, false
	}
	event := &sdkevent.MarketplaceBuy{Buyer: hub.player(buyerID), ListingID: listingID, SellerID: sellerID, BuyerPrice: buyerPrice}
	err := hub.Dispatch(ctx, event)
	return event.BuyerPrice, errors.Is(err, ErrEventCancelled)
}

// DispatchCraftingCraft sends one cancellable crafting transaction.
func (hub *Hub) DispatchCraftingCraft(ctx context.Context, playerID int64, recipeID int64, rewardDefinitionID int64) (int64, bool) {
	if !hub.HasListeners(sdkevent.CraftingCraftName) {
		return rewardDefinitionID, false
	}
	event := &sdkevent.CraftingCraft{Player: hub.player(playerID), RecipeID: recipeID, RewardDefinitionID: rewardDefinitionID}
	err := hub.Dispatch(ctx, event)
	return event.RewardDefinitionID, errors.Is(err, ErrEventCancelled)
}
