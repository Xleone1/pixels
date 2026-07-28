package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// fakeChecker returns one permission decision.
type fakeChecker struct {
	// allowed stores the returned decision.
	allowed bool
	// err stores the returned failure.
	err error
	// calls stores permission lookup count.
	calls int
}

// currencyEventsForTest mutates or cancels one grant.
type currencyEventsForTest struct {
	// amount stores the replacement delta.
	amount int64
	// cancelled stores the veto decision.
	cancelled bool
	// calls stores dispatch count.
	calls int
}

// DispatchCurrencyGrant returns the configured plugin decision.
func (events *currencyEventsForTest) DispatchCurrencyGrant(_ context.Context, player sdkplayer.Player, currencyType int32, _ int64, actor string) (int64, bool) {
	events.calls++
	if player.ID != 7 || currencyType != -1 || actor != ActorPlugin {
		return 0, true
	}
	return events.amount, events.cancelled
}

// HasPermission resolves one fixture permission decision.
func (checker *fakeChecker) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	checker.calls++
	if node != currency.InfiniteBalance {
		return false, errors.New("unexpected permission node")
	}
	return checker.allowed, checker.err
}

// TestInfiniteBalanceSkipsPlayerDeductions verifies privileged purchases remain free.
func TestInfiniteBalanceSkipsPlayerDeductions(t *testing.T) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 25}}}
	checker := &fakeChecker{allowed: true}
	service := newTestService(t, store)
	service.permissions = checker

	amount, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer})
	if err != nil || amount != 25 {
		t.Fatalf("unexpected balance=%d err=%v", amount, err)
	}
	if store.mutation.Amount != 0 || checker.calls != 1 {
		t.Fatalf("expected no persisted deduction, mutation=%#v calls=%d", store.mutation, checker.calls)
	}
}

// TestInfiniteBalanceDoesNotBypassAdminDeductions verifies actor scoping.
func TestInfiniteBalanceDoesNotBypassAdminDeductions(t *testing.T) {
	store := &fakeStore{grantBalance: currencymodel.Balance{Amount: 20}}
	checker := &fakeChecker{allowed: true}
	service := newTestService(t, store)
	service.permissions = checker

	amount, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorAdmin})
	if err != nil || amount != 20 || store.mutation.Amount != -5 || checker.calls != 0 {
		t.Fatalf("unexpected balance=%d mutation=%#v calls=%d err=%v", amount, store.mutation, checker.calls, err)
	}
}

// TestInfiniteBalancePropagatesPermissionFailures verifies purchase checks fail closed.
func TestInfiniteBalancePropagatesPermissionFailures(t *testing.T) {
	failure := errors.New("permission unavailable")
	service := newTestService(t, &fakeStore{})
	service.permissions = &fakeChecker{err: failure}

	_, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer})
	if !errors.Is(err, failure) {
		t.Fatalf("expected permission failure, got %v", err)
	}
}

// TestPluginCurrencyGrantMutatesAndCancelsBeforePersistence verifies the grant seam is transactional.
func TestPluginCurrencyGrantMutatesAndCancelsBeforePersistence(t *testing.T) {
	t.Run("mutates", func(t *testing.T) {
		store := &fakeStore{grantBalance: currencymodel.Balance{Amount: 12}}
		events := &currencyEventsForTest{amount: 12}
		service := newTestService(t, store)
		service.SetPluginRuntime(events)

		amount, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: 5, ActorKind: ActorPlugin, Reason: "plugin:test"})
		if err != nil || amount != 12 || store.mutation.Amount != 12 || store.mutation.ActorKind != ActorPlugin || store.mutation.Reason != "plugin:test" || !store.mutation.Ledger || events.calls != 1 {
			t.Fatalf("amount=%d mutation=%#v calls=%d err=%v", amount, store.mutation, events.calls, err)
		}
	})
	t.Run("cancels", func(t *testing.T) {
		store := &fakeStore{}
		events := &currencyEventsForTest{cancelled: true}
		service := newTestService(t, store)
		service.SetPluginRuntime(events)

		_, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: 5, ActorKind: ActorPlugin, Reason: "plugin:test"})
		if !errors.Is(err, ErrCancelledByPlugin) || store.mutation.PlayerID != 0 || events.calls != 1 {
			t.Fatalf("mutation=%#v calls=%d err=%v", store.mutation, events.calls, err)
		}
	})
}

// BenchmarkInfiniteBalanceDeduction measures permission-aware free purchases.
func BenchmarkInfiniteBalanceDeduction(b *testing.B) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 25}}}
	service := newTestService(b, store)
	service.permissions = &fakeChecker{allowed: true}
	params := GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		amount, err := service.Grant(ctx, params)
		if err != nil || amount != 25 {
			b.Fatalf("unexpected amount=%d err=%v", amount, err)
		}
	}
}
