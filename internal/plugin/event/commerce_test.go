package event

import (
	"context"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	"go.uber.org/zap"
)

// TestCommerceDispatchersApplyMutationsAndCancellation verifies both commerce gates.
func TestCommerceDispatchersApplyMutationsAndCancellation(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	hub.SetPlayerFinder(playerFinder{found: true})
	scope := pluginruntime.NewScope("commerce")
	_ = hub.listen(scope, sdkevent.FurniturePlaceName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.FurniturePlace)
		event.X, event.Y, event.Z, event.Rotation = 4, 5, 1.5, 6
		event.WallPosition = ":w=1,1 l=2,2"
		event.SetCancelled(true)
		return nil
	})
	params, cancelled := hub.DispatchFurniturePlace(context.Background(), furnitureservice.PlaceParams{
		ItemID: 2, ActorPlayerID: 7, RoomID: 3,
		Placement: furnituremodel.Placement{X: 1, Y: 2, Rotation: 2},
	})
	if !cancelled || params.Placement.X != 4 || params.Placement.Y != 5 || params.Placement.Z != 1.5 ||
		params.Placement.Rotation != 6 || params.WallPosition == "" {
		t.Fatalf("params=%#v cancelled=%v", params, cancelled)
	}
	_ = hub.listen(scope, sdkevent.FurnitureMoveName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.FurnitureMove)
		event.X = 12
		return nil
	})
	moved, cancelled := hub.DispatchFurnitureMove(context.Background(), furnitureservice.MoveParams{
		ItemID: 2, ActorPlayerID: 7, RoomID: 3, Placement: furnituremodel.Placement{X: 1, Y: 2, Rotation: 2},
	})
	if cancelled || moved.Placement.X != 12 {
		t.Fatalf("move=%#v cancelled=%v", moved, cancelled)
	}
	_ = hub.listen(scope, sdkevent.FurniturePickupName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		current.(sdkevent.Cancellable).SetCancelled(true)
		return nil
	})
	if !hub.DispatchFurniturePickup(context.Background(), furnitureservice.PickupParams{ItemID: 2, ActorPlayerID: 7, RoomID: 3}) {
		t.Fatal("pickup was not cancelled")
	}

	_ = hub.listen(scope, sdkevent.CatalogPurchaseName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		event := current.(*sdkevent.CatalogPurchase)
		event.CreditsCost, event.PointsCost, event.PointsType = 0, 2, 7
		event.SetCancelled(true)
		return nil
	})
	credits, points, pointsType, cancelled := hub.DispatchCatalogPurchase(context.Background(), 7, catalogmodel.Item{
		CostCredits: 10, CostPoints: 20, PointsType: 5,
	}, 1)
	if credits != 0 || points != 2 || pointsType != 7 || !cancelled {
		t.Fatalf("prices=%d/%d/%d cancelled=%v", credits, points, pointsType, cancelled)
	}
	_ = hub.listen(scope, sdkevent.MarketplaceBuyName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		current.(*sdkevent.MarketplaceBuy).BuyerPrice = 15
		return nil
	})
	price, cancelled := hub.DispatchMarketplaceBuy(context.Background(), 7, 8, 9, 20)
	if cancelled || price != 15 {
		t.Fatalf("price=%d cancelled=%v", price, cancelled)
	}
	_ = hub.listen(scope, sdkevent.CraftingCraftName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		current.(*sdkevent.CraftingCraft).RewardDefinitionID = 55
		return nil
	})
	reward, cancelled := hub.DispatchCraftingCraft(context.Background(), 7, 10, 11)
	if cancelled || reward != 55 {
		t.Fatalf("reward=%d cancelled=%v", reward, cancelled)
	}
}

// BenchmarkCommerceDispatchersWithoutListeners measures original event guard allocations.
func BenchmarkCommerceDispatchersWithoutListeners(b *testing.B) {
	hub := NewHub(time.Second, zap.NewNop())
	ctx := context.Background()
	place := furnitureservice.PlaceParams{ActorPlayerID: 7}
	item := catalogmodel.Item{CostCredits: 10}
	b.Run("furniture_place", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = hub.DispatchFurniturePlace(ctx, place)
		}
	})
	b.Run("catalog_purchase", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _, _, _ = hub.DispatchCatalogPurchase(ctx, 7, item, 1)
		}
	})
}
