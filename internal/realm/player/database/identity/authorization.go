package identity

import (
	"context"
	"fmt"

	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	"github.com/niflaot/pixels/pkg/postgres"
)

// SetAuthorization replaces and audits one self-service rename authorization.
func (repository *Repository) SetAuthorization(ctx context.Context, playerID int64, allowed bool, actorPlayerID int64, reason string) error {
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		tag, updateErr := executor.Exec(txCtx, `update player_profiles set allow_name_change=$2,updated_at=now(),version=version+1 where player_id=$1`, playerID, allowed)
		if updateErr != nil {
			return updateErr
		}
		if tag.RowsAffected() == 0 {
			return playeridentity.ErrPlayerNotFound
		}
		_, insertErr := executor.Exec(txCtx, `insert into player_name_change_authorizations(player_id,allowed,actor_player_id,reason) values($1,$2,$3,$4)`, playerID, allowed, actorPlayerID, reason)
		return insertErr
	})
	if err != nil {
		return fmt.Errorf("set player name change authorization: %w", err)
	}
	return nil
}
