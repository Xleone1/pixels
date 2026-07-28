package core

import (
	"context"
	"errors"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"testing"
)

// settlementStore executes work without persistence for focused settlement tests.
type settlementStore struct {
	// transactions counts settlement transaction attempts.
	transactions int
	// err rejects settlement before transaction work.
	err error
}

func (store *settlementStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	store.transactions++
	if store.err != nil {
		return store.err
	}
	return work(ctx)
}
func (*settlementStore) InsertAudit(context.Context, traderecord.Audit) error { return nil }
func (*settlementStore) ListAudits(context.Context, int64, int32) ([]traderecord.Audit, error) {
	return nil, nil
}

// settlementFurniture provides stable owned inventory items.
type settlementFurniture struct {
	// transfers counts ordinary ownership mutations.
	transfers int
	// deletions counts redeemed furniture consumption.
	deletions int
}

func (*settlementFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, AllowTrade: true}, true, nil
}
func (*settlementFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}
func (*settlementFurniture) FindItemByID(_ context.Context, itemID int64) (furnituremodel.Item, bool, error) {
	owner := int64(1)
	if itemID == 20 {
		owner = 2
	}
	return furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: itemID}}, DefinitionID: 1, OwnerPlayerID: owner}, true, nil
}
func (*settlementFurniture) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}
func (*settlementFurniture) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}
func (*settlementFurniture) ReserveForMarketplace(context.Context, int64, int64) (furnituremodel.Item, furnituremodel.Definition, error) {
	return furnituremodel.Item{}, furnituremodel.Definition{}, nil
}
func (*settlementFurniture) ReleaseFromMarketplace(context.Context, int64, int64) error { return nil }
func (*settlementFurniture) TransferFromMarketplace(context.Context, int64, int64, int64) error {
	return nil
}
func (furniture *settlementFurniture) TransferInventoryItem(context.Context, int64, int64, int64) error {
	furniture.transfers++
	return nil
}
func (furniture *settlementFurniture) DeleteInventoryItem(context.Context, int64, int64) error {
	furniture.deletions++
	return nil
}

// settlementCurrency accepts settlement grants.
type settlementCurrency struct {
	// grants counts currency mutations.
	grants int
}

func (currency *settlementCurrency) Grant(context.Context, currencyservice.GrantParams) (int64, error) {
	currency.grants++
	return 0, nil
}

// TestSettleRevalidatesAndTransfersBothOffers verifies the happy settlement path.
func TestSettleRevalidatesAndTransfersBothOffers(t *testing.T) {
	furniture := &settlementFurniture{}
	service := &Service{config: Options{}, store: &settlementStore{}, furniture: furniture, currencies: &settlementCurrency{}}
	session := &traderuntime.Session{RoomID: 9, First: traderuntime.Participant{PlayerID: 1, Items: []int64{10}}, Second: traderuntime.Participant{PlayerID: 2, Items: []int64{20}}}
	if err := service.settle(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	if furniture.transfers != 2 || furniture.deletions != 0 {
		t.Fatalf("transfers=%d deletions=%d", furniture.transfers, furniture.deletions)
	}
}

// TestSettlementFailureAlwaysClosesTheSession verifies optional cancellation hooks cannot retain an invalid trade.
func TestSettlementFailureAlwaysClosesTheSession(t *testing.T) {
	fixture := newStartFixture(t, false, nil)
	settlementErr := errors.New("settlement failed")
	fixture.service.store = &settlementStore{err: settlementErr}
	fixture.service.furniture = &settlementFurniture{}
	fixture.service.currencies = &settlementCurrency{}
	fixture.service.SetPluginRuntime(tradeEventsForTest{cancel: true})
	session, err := fixture.service.Start(context.Background(), 1, fixture.units[2], "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	session.SetAccepted(1, true)
	session.SetAccepted(2, true)
	if _, err = fixture.service.Confirm(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	completed, err := fixture.service.Confirm(context.Background(), 2)
	if completed || !errors.Is(err, settlementErr) || fixture.service.registry.ActiveCount() != 0 {
		t.Fatalf("completed=%v active=%d err=%v", completed, fixture.service.registry.ActiveCount(), err)
	}
}

// BenchmarkSettle measures validation and atomic ownership settlement coordination.
func BenchmarkSettle(b *testing.B) {
	service := &Service{config: Options{}, store: &settlementStore{}, furniture: &settlementFurniture{}, currencies: &settlementCurrency{}}
	session := &traderuntime.Session{RoomID: 9, First: traderuntime.Participant{PlayerID: 1, Items: []int64{10}}, Second: traderuntime.Participant{PlayerID: 2, Items: []int64{20}}}
	b.ReportAllocs()
	for b.Loop() {
		if err := service.settle(context.Background(), session); err != nil {
			b.Fatal(err)
		}
	}
}
