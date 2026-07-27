package service

import (
	"context"
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogpricing "github.com/niflaot/pixels/internal/realm/catalog/pricing"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestPageProjectsFreePricesWithoutMutatingTheCache verifies reversible debug pricing.
func TestPageProjectsFreePricesWithoutMutatingTheCache(t *testing.T) {
	item := dynamicItem(10, 1)
	item.CostCredits = 10
	item.CostPoints = 4
	store := &fakeStore{pages: []catalogmodel.Page{dynamicPage(1, catalogmodel.DefaultLayout)}, items: []catalogmodel.Item{item}}
	service := dynamicService(t, store)
	pricing := catalogpricing.New(true)
	service.WithPricing(pricing)

	_, free, err := service.Page(context.Background(), 1, 7, false)
	if err != nil || len(free) != 1 || free[0].CostCredits != 0 || free[0].CostPoints != 0 {
		t.Fatalf("free=%#v error=%v", free, err)
	}
	pricing.ToggleFreeItems()
	_, paid, err := service.Page(context.Background(), 1, 7, false)
	if err != nil || len(paid) != 1 || paid[0].CostCredits != 10 || paid[0].CostPoints != 4 {
		t.Fatalf("paid=%#v error=%v", paid, err)
	}
}

// TestRecentPurchasesPageUsesPlayerHistory verifies dynamic offer ordering and visibility.
func TestRecentPurchasesPageUsesPlayerHistory(t *testing.T) {
	pages := []catalogmodel.Page{
		dynamicPage(1, catalogmodel.DefaultLayout),
		dynamicPage(2, catalogmodel.RecentPurchasesLayout),
	}
	first := dynamicItem(10, 1)
	second := dynamicItem(11, 1)
	club := dynamicItem(12, 1)
	club.ClubOnly = true
	store := &fakeStore{pages: pages, items: []catalogmodel.Item{first, second, club}, recentItemIDs: []int64{11, 999, 12, 10}}
	service := dynamicService(t, store)

	_, regular, err := service.Page(context.Background(), 2, 7, false)
	if err != nil || len(regular) != 2 || regular[0].ID != 11 || regular[1].ID != 10 {
		t.Fatalf("regular=%#v error=%v", regular, err)
	}
	_, entitled, err := service.Page(context.Background(), 2, 7, true)
	if err != nil || len(entitled) != 3 || entitled[1].ID != 12 {
		t.Fatalf("entitled=%#v error=%v", entitled, err)
	}
}

// TestSoldLimitedItemsPageIncludesDisabledCompletedSeries verifies dynamic LTD archival.
func TestSoldLimitedItemsPageIncludesDisabledCompletedSeries(t *testing.T) {
	pages := []catalogmodel.Page{
		dynamicPage(1, catalogmodel.DefaultLayout),
		dynamicPage(2, catalogmodel.SoldLimitedItemsLayout),
	}
	sold := dynamicItem(10, 1)
	sold.LimitedStack, sold.LimitedSells, sold.Enabled = 5, 5, false
	available := dynamicItem(11, 1)
	available.LimitedStack, available.LimitedSells = 5, 4
	plain := dynamicItem(12, 1)
	store := &fakeStore{pages: pages, items: []catalogmodel.Item{available, plain, sold}}
	service := dynamicService(t, store)

	_, items, err := service.Page(context.Background(), 2, 7, false)
	if err != nil || len(items) != 1 || items[0].ID != sold.ID {
		t.Fatalf("items=%#v error=%v", items, err)
	}
}

// dynamicPage creates one accessible page fixture.
func dynamicPage(id int64, layout string) catalogmodel.Page {
	return catalogmodel.Page{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}},
		Name: "page", Layout: layout, Visible: true, Enabled: true,
	}
}

// dynamicItem creates one enabled furniture offer fixture.
func dynamicItem(id int64, pageID int64) catalogmodel.Item {
	return catalogmodel.Item{
		Base:   sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}},
		PageID: pageID, DefinitionID: id, RewardKind: catalogmodel.RewardFurniture,
		Name: "item", Amount: 1, PointsType: catalogmodel.CreditsType, Enabled: true,
	}
}

// dynamicService creates and refreshes a service for dynamic page tests.
func dynamicService(t *testing.T, store *fakeStore) *Service {
	t.Helper()
	definitions := make([]furnituremodel.Definition, 0, len(store.items))
	for _, item := range store.items {
		definitions = append(definitions, furnituremodel.Definition{
			Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: item.DefinitionID}},
			Name: item.Name,
		})
	}
	furniture := &fakeFurniture{definitions: definitions}
	service := New(store, &fakeCurrency{balances: make(map[int32]int64)}, furniture, nil, zap.NewNop())
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh dynamic service: %v", err)
	}
	return service
}
