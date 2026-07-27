package branding

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// testChecker returns one configured permission decision.
type testChecker struct {
	// allowed stores the permission decision.
	allowed bool
}

// HasPermission returns the configured permission decision.
func (checker testChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return checker.allowed, nil
}

// testStore records committed branding mutations.
type testStore struct {
	// mutation stores the latest upsert input.
	mutation Mutation
	// reason stores the latest disable reason.
	reason string
}

// List returns no test configurations.
func (store *testStore) List(context.Context, int64) ([]Config, error) { return nil, nil }

// ListCompatible returns no test furniture.
func (store *testStore) ListCompatible(context.Context, int64) ([]CompatibleItem, error) {
	return nil, nil
}

// Upsert records and returns one committed configuration.
func (store *testStore) Upsert(_ context.Context, mutation Mutation, extraData string) (Projection, error) {
	store.mutation = mutation
	return Projection{
		Config: Config{
			ID: 1, RoomID: mutation.RoomID, FurnitureItemID: mutation.FurnitureItemID,
			Kind: mutation.Kind, ImageURL: mutation.ImageURL, ClickURL: mutation.ClickURL,
			AssetRef: mutation.AssetRef, Enabled: true, Version: 1,
		},
		ExtraData: extraData,
	}, nil
}

// Disable records and returns one disabled configuration.
func (store *testStore) Disable(_ context.Context, roomID int64, brandingID int64, _ int64, _ int64, reason string, extraData string) (Projection, error) {
	store.reason = reason
	return Projection{Config: Config{ID: brandingID, RoomID: roomID, Version: 2}, ExtraData: extraData}, nil
}

// newTestService creates an isolated branding service.
func newTestService(store Store, allowed bool) *Service {
	return New(
		store,
		testChecker{allowed: allowed},
		permission.Node("room.branding.manage"),
		roomlive.NewRegistry(nil),
		netconn.NewRegistry(),
	)
}

// TestUpsertNormalizesURLsAndAuditCopy verifies canonical persistence input.
func TestUpsertNormalizesURLsAndAuditCopy(t *testing.T) {
	store := &testStore{}
	service := newTestService(store, true)
	_, err := service.Upsert(context.Background(), Mutation{
		RoomID: 4, FurnitureItemID: 8, Kind: KindBillboard,
		ImageURL: " https://cdn.example/banner.png ", ClickURL: " https://example.com ",
		AssetRef: " asset:9 ", ActorPlayerID: 2, Reason: " campaign launch ",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if store.mutation.ImageURL != "https://cdn.example/banner.png" ||
		store.mutation.ClickURL != "https://example.com" ||
		store.mutation.AssetRef != "asset:9" ||
		store.mutation.Reason != "campaign launch" {
		t.Fatalf("mutation was not normalized: %#v", store.mutation)
	}
}

// TestUpsertRejectsBackgroundClickAndForbiddenActor verifies safety gates.
func TestUpsertRejectsBackgroundClickAndForbiddenActor(t *testing.T) {
	valid := Mutation{
		RoomID: 4, FurnitureItemID: 8, Kind: KindBackground,
		ImageURL: "https://cdn.example/banner.png", AssetRef: "asset:9",
		ActorPlayerID: 2, Reason: "launch",
	}
	service := newTestService(&testStore{}, true)
	valid.ClickURL = "https://example.com"
	if _, err := service.Upsert(context.Background(), valid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected invalid background click, got %v", err)
	}
	valid.ClickURL = ""
	service = newTestService(&testStore{}, false)
	if _, err := service.Upsert(context.Background(), valid); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden actor, got %v", err)
	}
}

// TestDisableRequiresVersionAndNormalizesReason verifies safe disable input.
func TestDisableRequiresVersionAndNormalizesReason(t *testing.T) {
	store := &testStore{}
	service := newTestService(store, true)
	if _, err := service.Disable(context.Background(), 4, 1, 0, 2, "cleanup"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected invalid version, got %v", err)
	}
	if _, err := service.Disable(context.Background(), 4, 1, 1, 2, " cleanup "); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if store.reason != "cleanup" {
		t.Fatalf("reason=%q", store.reason)
	}
}

// TestReadRequiresAuthorizedActor verifies that branding metadata is not an unattributed read.
func TestReadRequiresAuthorizedActor(t *testing.T) {
	service := newTestService(&testStore{}, false)
	if _, err := service.List(context.Background(), 4, 2); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden list, got %v", err)
	}
	if _, err := service.ListCompatible(context.Background(), 4, 0); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected invalid actor, got %v", err)
	}
}
