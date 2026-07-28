package economy

import (
	"context"
	"errors"
	"testing"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// economyStoreForTest records plugin-authored mutations.
type economyStoreForTest struct {
	// mutation stores the last durable request.
	mutation currencyrepo.Mutation
	// balance stores the current fixture value.
	balance int64
}

// FindBalance returns the current fixture balance.
func (store *economyStoreForTest) FindBalance(_ context.Context, playerID int64, currencyType int32) (currencymodel.Balance, bool, error) {
	return currencymodel.Balance{PlayerID: playerID, CurrencyType: currencyType, Amount: store.balance}, true, nil
}

// ListBalances returns the current fixture balance.
func (store *economyStoreForTest) ListBalances(context.Context, int64) ([]currencymodel.Balance, error) {
	return nil, nil
}

// Grant records and applies one signed mutation.
func (store *economyStoreForTest) Grant(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	store.mutation = mutation
	store.balance += mutation.Amount
	return currencyrepo.Result{Balance: currencymodel.Balance{PlayerID: mutation.PlayerID, CurrencyType: mutation.CurrencyType, Amount: store.balance}, Delta: mutation.Amount}, nil
}

// Set records and applies one absolute mutation.
func (store *economyStoreForTest) Set(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	store.mutation = mutation
	delta := mutation.Amount - store.balance
	store.balance = mutation.Amount
	return currencyrepo.Result{Balance: currencymodel.Balance{PlayerID: mutation.PlayerID, CurrencyType: mutation.CurrencyType, Amount: store.balance}, Delta: delta}, nil
}

// economyServiceForTest creates a ledger-enabled currency manager.
func economyServiceForTest(t *testing.T, store *economyStoreForTest) *currencyservice.Service {
	t.Helper()
	catalog, err := currency.NewCatalog([]currencymodel.Definition{{Type: -1, Key: "credits", Ledger: true}}, []int32{-1})
	if err != nil {
		t.Fatal(err)
	}
	return currencyservice.New(store, catalog, nil, nil)
}

// TestEconomyAccessUsesPluginAuditIdentity verifies agency stays inside the real currency service.
func TestEconomyAccessUsesPluginAuditIdentity(t *testing.T) {
	store := &economyStoreForTest{balance: 10}
	access := NewAccess(economyServiceForTest(t, store), pluginruntime.NewScope("rewards"))
	balance, err := access.Grant(7, -1, 5)
	if err != nil || balance != 15 {
		t.Fatalf("balance=%d err=%v", balance, err)
	}
	if store.mutation.ActorKind != currencyservice.ActorPlugin || store.mutation.Reason != "plugin:rewards:grant" || !store.mutation.Ledger {
		t.Fatalf("mutation=%#v", store.mutation)
	}
}

// TestEconomyAccessReadsSetsAndListsTypes verifies the remaining bounded economy actions.
func TestEconomyAccessReadsSetsAndListsTypes(t *testing.T) {
	store := &economyStoreForTest{balance: 10}
	access := NewAccess(economyServiceForTest(t, store), pluginruntime.NewScope("rewards"))
	balance, err := access.Set(7, -1, 25)
	if err != nil || balance != 25 {
		t.Fatalf("set balance=%d err=%v", balance, err)
	}
	if store.mutation.ActorKind != currencyservice.ActorPlugin || store.mutation.Reason != "plugin:rewards:set" {
		t.Fatalf("mutation=%#v", store.mutation)
	}
	balance, err = access.Balance(7, -1)
	if err != nil || balance != 25 {
		t.Fatalf("read balance=%d err=%v", balance, err)
	}
	definitions, err := access.Types()
	if err != nil || len(definitions) != 1 || definitions[0].Type != -1 || definitions[0].Key != "credits" || !definitions[0].Ledger {
		t.Fatalf("definitions=%#v err=%v", definitions, err)
	}
}

// TestEconomyAccessFailsClosedForDisabledScope verifies late callbacks cannot mutate balances.
func TestEconomyAccessFailsClosedForDisabledScope(t *testing.T) {
	store := &economyStoreForTest{balance: 10}
	scope := pluginruntime.NewScope("disabled")
	scope.Disable()
	access := NewAccess(economyServiceForTest(t, store), scope)
	if _, err := access.Grant(7, -1, 5); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if store.balance != 10 {
		t.Fatalf("balance changed to %d", store.balance)
	}
}
