package identity

import (
	"context"
	"fmt"

	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
)

// NameChanges returns recent username audit entries newest first.
func (repository *Repository) NameChanges(ctx context.Context, playerID int64, limit int) ([]playeridentity.NameChange, error) {
	rows, err := repository.pool.Query(ctx, `select id,player_id,old_username,new_username,actor_player_id,reason,source,changed_at from player_name_changes where player_id=$1 order by changed_at desc,id desc limit $2`, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("query player name changes: %w", err)
	}
	defer rows.Close()

	changes := make([]playeridentity.NameChange, 0)
	for rows.Next() {
		var change playeridentity.NameChange
		if err = rows.Scan(&change.ID, &change.PlayerID, &change.OldUsername, &change.NewUsername, &change.ActorPlayerID, &change.Reason, &change.Source, &change.ChangedAt); err != nil {
			return nil, fmt.Errorf("scan player name change: %w", err)
		}
		changes = append(changes, change)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player name changes: %w", err)
	}
	return changes, nil
}
