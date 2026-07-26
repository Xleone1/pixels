// Package identity implements PostgreSQL atomic player renames.
package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	playerconstraint "github.com/niflaot/pixels/internal/realm/player/database/constraint"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository commits player identity replacements.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates an identity repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Rename commits one one-shot username replacement and audit row.
func (repository *Repository) Rename(ctx context.Context, playerID int64, username string, changedAt time.Time, cooldown time.Duration) (result playeridentity.RenameResult, err error) {
	return repository.rename(ctx, playerID, username, playerID, "self-service", "client", changedAt, cooldown, true)
}

// RenameAdmin commits one attributed administrative username replacement.
func (repository *Repository) RenameAdmin(ctx context.Context, playerID int64, username string, actorPlayerID int64, reason string, changedAt time.Time, cooldown time.Duration) (result playeridentity.RenameResult, err error) {
	return repository.rename(ctx, playerID, username, actorPlayerID, reason, "api", changedAt, cooldown, false)
}

// rename commits one username replacement and its audit entry.
func (repository *Repository) rename(ctx context.Context, playerID int64, username string, actorPlayerID int64, reason string, source string, changedAt time.Time, cooldown time.Duration, requireAvailable bool) (result playeridentity.RenameResult, err error) {
	err = postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var lastChangedAt pgtype.Timestamptz
		lockErr := executor.QueryRow(txCtx, `select p.username,pp.last_name_change_at from players p join player_profiles pp on pp.player_id=p.id where p.id=$1 and p.deleted_at is null for update of p,pp`, playerID).Scan(&result.OldUsername, &lastChangedAt)
		if errors.Is(lockErr, pgx.ErrNoRows) {
			return playeridentity.ErrPlayerNotFound
		}
		if lockErr == nil && requireAvailable && lastChangedAt.Valid && changedAt.Before(lastChangedAt.Time.Add(cooldown)) {
			return playeridentity.ErrRenameCooldown
		}
		if lockErr != nil {
			return lockErr
		}
		if _, updateErr := executor.Exec(txCtx, `update players set username=$2,updated_at=now(),version=version+1 where id=$1`, playerID, username); updateErr != nil {
			return updateErr
		}
		if _, updateErr := executor.Exec(txCtx, `update player_profiles set last_name_change_at=$2,updated_at=now(),version=version+1 where player_id=$1`, playerID, changedAt); updateErr != nil {
			return updateErr
		}
		if _, updateErr := executor.Exec(txCtx, `update rooms set owner_name=$2,updated_at=now(),version=version+1 where owner_player_id=$1 and deleted_at is null`, playerID, username); updateErr != nil {
			return updateErr
		}
		_, insertErr := executor.Exec(txCtx, `insert into player_name_changes(player_id,old_username,new_username,actor_player_id,reason,source,changed_at) values($1,$2,$3,$4,$5,$6,$7)`, playerID, result.OldUsername, username, actorPlayerID, reason, source, changedAt)
		return insertErr
	})
	if err != nil {
		if playerconstraint.IsActiveUsernameConflict(err) {
			return playeridentity.RenameResult{}, playeridentity.ErrUsernameTaken
		}
		return playeridentity.RenameResult{}, fmt.Errorf("rename player: %w", err)
	}
	result.NewUsername = username
	result.ChangedAt = changedAt
	result.AvailableAt = changedAt.Add(cooldown)
	return result, nil
}
