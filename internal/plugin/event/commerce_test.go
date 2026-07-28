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
}
