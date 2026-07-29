// Package tests exercises Messenger plugin guards outside the six-pair core package.
package tests

import (
	"context"
	"errors"
	"testing"

	messengercore "github.com/niflaot/pixels/internal/realm/messenger/core"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// messengerStore supplies request reads and records unexpected writes.
type messengerStore struct {
	messengermodel.Store
	// created counts friend-request writes.
	created int
	// accepted counts friendship writes.
	accepted int
}

// IsFriend reports no existing friendship.
func (*messengerStore) IsFriend(context.Context, int64, int64) (bool, error) { return false, nil }

// HasRequestEither reports no pending request.
func (*messengerStore) HasRequestEither(context.Context, int64, int64) (bool, error) {
	return false, nil
}

// CountFriends reports an empty friend list.
func (*messengerStore) CountFriends(context.Context, int64) (int, error) { return 0, nil }

// CreateRequest records one unexpected request write.
func (store *messengerStore) CreateRequest(context.Context, int64, int64) (bool, error) {
	store.created++
	return true, nil
}

// AcceptRequest records one unexpected friendship write.
func (store *messengerStore) AcceptRequest(context.Context, int64, int64) (bool, error) {
	store.accepted++
	return true, nil
}

// messengerPlayers resolves two deterministic players.
type messengerPlayers struct{ playerservice.Manager }

// FindByID returns one deterministic player.
func (*messengerPlayers) FindByID(_ context.Context, id int64) (playerservice.Record, bool, error) {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, Username: map[int64]string{1: "demo", 2: "alice"}[id]}}, id == 1 || id == 2, nil
}

// FindByUsername returns the deterministic friend-request target.
func (players *messengerPlayers) FindByUsername(ctx context.Context, _ string) (playerservice.Record, bool, error) {
	return players.FindByID(ctx, 2)
}

// messengerEvents vetoes every friendship mutation.
type messengerEvents struct{}

// DispatchFriendRequest vetoes request creation.
func (messengerEvents) DispatchFriendRequest(context.Context, int64, int64) bool { return true }

// DispatchFriendAccept vetoes friendship acceptance.
func (messengerEvents) DispatchFriendAccept(context.Context, int64, int64) bool { return true }

// TestPluginCancellationPrecedesFriendshipPersistence verifies both guards.
func TestPluginCancellationPrecedesFriendshipPersistence(t *testing.T) {
	store := &messengerStore{}
	service := messengercore.New(messengercore.Options{MaxFriends: 100, MaxFriendsClub: 100}, store, &messengerPlayers{}, nil, nil, nil, nil, nil, messengercore.Nodes{}, nil)
	service.SetPluginRuntime(messengerEvents{})
	if _, err := service.SendRequest(context.Background(), 1, "alice"); !errors.Is(err, messengercore.ErrCancelledByPlugin) {
		t.Fatalf("request error=%v", err)
	}
	if _, err := service.Accept(context.Background(), 1, 2); !errors.Is(err, messengercore.ErrCancelledByPlugin) {
		t.Fatalf("accept error=%v", err)
	}
	if store.created != 0 || store.accepted != 0 {
		t.Fatalf("created=%d accepted=%d", store.created, store.accepted)
	}
}
