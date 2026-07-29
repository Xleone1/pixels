// Package tests exercises membership plugin guards outside the six-pair core package.
package tests

import (
	"context"
	"errors"
	"testing"

	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	"github.com/niflaot/pixels/internal/realm/group/membership"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// membershipStore exposes only authorization reads and records unexpected writes.
type membershipStore struct {
	grouprecord.Store
	// actor stores the authorized roster manager.
	actor grouprecord.Membership
}

// Membership returns the configured roster manager.
func (store *membershipStore) Membership(context.Context, int64, int64) (grouprecord.Membership, bool, error) {
	return store.actor, true, nil
}

// membershipEvents vetoes every membership mutation.
type membershipEvents struct{}

// DispatchGroupMembershipChange vetoes one membership mutation.
func (membershipEvents) DispatchGroupMembershipChange(context.Context, string, int64, int64, string) bool {
	return true
}

// TestPluginCancellationPrecedesMembershipPersistence verifies all three guards.
func TestPluginCancellationPrecedesMembershipPersistence(t *testing.T) {
	store := &membershipStore{actor: grouprecord.Membership{GroupID: 3, PlayerID: 1, Role: grouprecord.Owner}}
	service := membership.New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	service.SetPluginRuntime(membershipEvents{})
	if _, _, err := service.Join(context.Background(), 7, 3); !errors.Is(err, membership.ErrCancelledByPlugin) {
		t.Fatalf("join error=%v", err)
	}
	if _, _, err := service.Add(context.Background(), 1, 3, 7, grouprecord.Member); !errors.Is(err, membership.ErrCancelledByPlugin) {
		t.Fatalf("add error=%v", err)
	}
	if _, err := service.Accept(context.Background(), 1, 3, 7); !errors.Is(err, membership.ErrCancelledByPlugin) {
		t.Fatalf("accept error=%v", err)
	}
}
