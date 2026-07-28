package service

import (
	"context"
	"errors"
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogpricing "github.com/niflaot/pixels/internal/realm/catalog/pricing"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// catalogEventsForTest mutates or cancels one purchase.
type catalogEventsForTest struct {
	// listening reports an active purchase listener.
	listening bool
	// credits stores the replacement credit price.
	credits int64
	// cancelled stores the veto decision.
	cancelled bool
	// calls counts dispatched purchase attempts.
	calls *int
}

// HasListeners reports the configured purchase listener state.
func (events catalogEventsForTest) HasListeners(name string) bool {
	return events.listening && name == "catalog.purchase"
}

// DispatchCatalogPurchase returns the configured plugin decision.
func (events catalogEventsForTest) DispatchCatalogPurchase(_ context.Context, _ int64, item catalogmodel.Item, _ int32) (int64, int64, int32, bool) {
	if events.calls != nil {
		(*events.calls)++
	}
	return events.credits, item.CostPoints, item.PointsType, events.cancelled
}

// giftPlayerFinder resolves one configured recipient fixture.
type giftPlayerFinder struct {
	// record stores the configured recipient.
	record playerservice.Record
}

// FindByID supplies unused player lookup behavior.
func (finder giftPlayerFinder) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// FindByUsername resolves the configured gift recipient.
func (finder giftPlayerFinder) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	return finder.record, username == finder.record.Player.Username, nil
}

// TestPurchaseGiftChargesBuyerAndWrapsRecipientItem verifies durable gift ownership and metadata.
func TestPurchaseGiftChargesBuyerAndWrapsRecipientItem(t *testing.T) {
	item := itemForTest()
	item.Giftable = true
	fixture := newServiceFixture(t, item)
	receiver := playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 12}}, Username: "alice"}
	fixture.service.WithPlayers(giftPlayerFinder{record: playerservice.Record{Player: receiver}})

	result, err := fixture.service.PurchaseGift(context.Background(), GiftPurchaseParams{
		BuyerID: 7, ReceiverName: " alice ", CatalogItemID: item.ID, SpriteID: 3372, BoxID: 8, RibbonID: 10,
		Message: "Enjoy", ShowMyFace: true,
	})
	if err != nil || result.RecipientPlayerID != 12 || len(result.GrantedItems) != 1 {
		t.Fatalf("unexpected gift result %#v error %v", result, err)
	}
	if len(fixture.currency.calls) != 1 || fixture.currency.calls[0].PlayerID != 7 {
		t.Fatalf("unexpected buyer charge %#v", fixture.currency.calls)
	}
	if len(fixture.furniture.giftCalls) != 1 {
		t.Fatalf("unexpected gift grants %#v", fixture.furniture.giftCalls)
	}
	grant := fixture.furniture.giftCalls[0]
	if grant.OwnerPlayerID != 12 || grant.SpriteID != 3372 || grant.BoxID != 8 || grant.RibbonID != 10 || grant.SenderPlayerID == nil || *grant.SenderPlayerID != 7 {
		t.Fatalf("unexpected gift metadata %#v", grant)
	}
}

// TestPluginCatalogPurchasePreservesNativePricingAndVetoesBeforeCommit verifies purchase interception.
func TestPluginCatalogPurchasePreservesNativePricingAndVetoesBeforeCommit(t *testing.T) {
	t.Run("mutates", func(t *testing.T) {
		fixture := newServiceFixture(t, itemForTest())
		fixture.service.SetPluginRuntime(catalogEventsForTest{listening: true, credits: 3})
		_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
		if err != nil || len(fixture.currency.calls) != 1 || fixture.currency.calls[0].Amount != -3 {
			t.Fatalf("calls=%#v err=%v", fixture.currency.calls, err)
		}
	})
	t.Run("zero price still grants", func(t *testing.T) {
		fixture := newServiceFixture(t, itemForTest())
		fixture.service.SetPluginRuntime(catalogEventsForTest{listening: true})
		result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
		if err != nil || len(fixture.currency.calls) != 0 || len(result.GrantedItems) != 1 {
			t.Fatalf("result=%#v calls=%#v err=%v", result, fixture.currency.calls, err)
		}
	})
	t.Run("cancels", func(t *testing.T) {
		fixture := newServiceFixture(t, itemForTest())
		fixture.service.SetPluginRuntime(catalogEventsForTest{listening: true, cancelled: true})
		_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
		if !errors.Is(err, ErrCancelledByPlugin) || fixture.store.txCalls != 0 || len(fixture.currency.calls) != 0 {
			t.Fatalf("transactions=%d calls=%#v err=%v", fixture.store.txCalls, fixture.currency.calls, err)
		}
	})
	t.Run("no listeners", func(t *testing.T) {
		fixture := newServiceFixture(t, itemForTest())
		fixture.service.WithPricing(catalogpricing.New(true))
		fixture.service.SetPluginRuntime(catalogEventsForTest{})
		result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
		if err != nil || result.ChargedCredits != 0 || len(fixture.currency.calls) != 0 {
			t.Fatalf("result=%#v calls=%#v err=%v", result, fixture.currency.calls, err)
		}
	})
	t.Run("native validation precedes interception", func(t *testing.T) {
		fixture := newServiceFixture(t, itemForTest())
		calls := 0
		fixture.service.SetPluginRuntime(catalogEventsForTest{listening: true, calls: &calls})
		_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Amount: 2})
		if !errors.Is(err, ErrInvalidAmount) || calls != 0 || fixture.store.txCalls != 0 {
			t.Fatalf("calls=%d transactions=%d err=%v", calls, fixture.store.txCalls, err)
		}
	})
}
