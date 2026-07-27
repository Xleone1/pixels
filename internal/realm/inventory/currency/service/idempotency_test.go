package service

import (
	"context"
	"testing"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

// TestGrantDoesNotRepublishIdempotentReplay verifies retries do not duplicate side effects.
func TestGrantDoesNotRepublishIdempotentReplay(t *testing.T) {
	store := &fakeStore{
		grantBalance: currencymodel.Balance{PlayerID: 7, CurrencyType: 5, Amount: 15},
		replayed:     true,
	}
	publisher := &fakePublisher{}
	service := newTestServiceWithPublisher(t, store, publisher)

	amount, err := service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: 5, Amount: 5, ActorKind: ActorAdmin,
		OperationKey: "1ee192e8-bcf4-48bf-aafb-81fcc95b8f17",
	})
	if err != nil || amount != 15 {
		t.Fatalf("unexpected replay amount=%d err=%v", amount, err)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected no replay event, got %#v", publisher.events)
	}
	if store.mutation.OperationKey == "" || store.mutation.RequestHash == "" {
		t.Fatalf("missing idempotency data %#v", store.mutation)
	}
}
