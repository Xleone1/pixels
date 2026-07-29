package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
)

// marketplaceCurrency records the last balance mutation.
type marketplaceCurrency struct {
	// mutation stores the last balance mutation.
	mutation currencyservice.GrantParams
}

// Grant records and accepts one balance mutation.
func (currency *marketplaceCurrency) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currency.mutation = params
	return params.Amount, nil
}

// marketplaceEvents supplies deterministic purchase interception.
type marketplaceEvents struct {
	// price replaces the buyer price when nonzero.
	price int64
	// cancelled vetoes the purchase.
	cancelled bool
}

// DispatchMarketplaceList preserves listing behavior for this fixture.
func (events marketplaceEvents) DispatchMarketplaceList(_ context.Context, _ int64, _ int64, price int64) (int64, bool) {
	return price, events.cancelled
}

// DispatchMarketplaceBuy returns the configured purchase mutation.
func (events marketplaceEvents) DispatchMarketplaceBuy(_ context.Context, _ int64, _ int64, _ int64, price int64) (int64, bool) {
	if events.price != 0 {
		price = events.price
	}
	return price, events.cancelled
}

// TestMarketplaceBuyPluginMutationAndCancellation verifies the transactional gate.
func TestMarketplaceBuyPluginMutationAndCancellation(t *testing.T) {
	store := &benchmarkStore{listing: marketrecord.Listing{ID: 3, SellerPlayerID: 8, FurnitureItemID: 17, RawPrice: 100, State: marketrecord.StateOpen, ExpiresAt: time.Now().Add(time.Hour)}}
	furniture := &benchmarkFurniture{}
	currency := &marketplaceCurrency{}
	service := marketcore.New(marketcore.Options{Enabled: true, CommissionPercent: 1}, store, furniture, currency, nil, nil)
	service.SetPluginRuntime(marketplaceEvents{price: 77})
	if _, err := service.Buy(context.Background(), 7, store.listing.ID); err != nil {
		t.Fatal(err)
	}
	if currency.mutation.Amount != -77 || !furniture.transferred || !store.marked {
		t.Fatalf("mutation=%#v transferred=%t marked=%t", currency.mutation, furniture.transferred, store.marked)
	}
	store.marked = false
	furniture.transferred = false
	currency.mutation = currencyservice.GrantParams{}
	service.SetPluginRuntime(marketplaceEvents{cancelled: true})
	if _, err := service.Buy(context.Background(), 7, store.listing.ID); !errors.Is(err, marketcore.ErrCancelledByPlugin) {
		t.Fatalf("expected plugin cancellation, got %v", err)
	}
	if currency.mutation.Amount != 0 || furniture.transferred || store.marked {
		t.Fatalf("mutation=%#v transferred=%t marked=%t", currency.mutation, furniture.transferred, store.marked)
	}
}
