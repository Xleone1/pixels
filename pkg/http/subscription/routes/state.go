package routes

import (
	"context"

	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// membershipSnapshot resolves one compact membership audit state.
func (dependencies Dependencies) membershipSnapshot(playerID int64) adminaction.Snapshot {
	return func(ctx context.Context) (any, error) {
		membership, _, _, found, err := dependencies.Subscriptions.Membership(ctx, playerID)
		if err != nil {
			return nil, err
		}
		if !found {
			return map[string]any{"active": false}, nil
		}
		return map[string]any{
			"active": true, "level": membership.Level,
			"expiresAt": membership.ExpiresAt, "version": membership.Version,
		}, nil
	}
}
