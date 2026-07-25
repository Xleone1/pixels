// Package profile implements PostgreSQL public-profile persistence.
package profile

import (
	"context"
	"fmt"
	"time"

	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists public tags and daily respect grants.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates a public-profile repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Tags returns ordered public tags.
func (repository *Repository) Tags(ctx context.Context, playerID int64) ([]string, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select tag from player_profile_tags where player_id=$1 order by position`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player profile tags: %w", err)
	}
	defer rows.Close()
	tags := make([]string, 0, playerprofile.MaxTags)
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan player profile tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// ReplaceTags atomically replaces ordered public tags.
func (repository *Repository) ReplaceTags(ctx context.Context, playerID int64, tags []string) error {
	return postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if _, err := executor.Exec(txCtx, `delete from player_profile_tags where player_id=$1`, playerID); err != nil {
			return err
		}
		for index, tag := range tags {
			if _, err := executor.Exec(txCtx, `insert into player_profile_tags(player_id,position,tag) values($1,$2,$3)`, playerID, index+1, tag); err != nil {
				return err
			}
		}
		return nil
	})
}

// RespectState returns total and remaining user and pet respect allowances.
func (repository *Repository) RespectState(ctx context.Context, playerID int64, date time.Time, userLimit int, petLimit int) (playerprofile.RespectState, error) {
	var state playerprofile.RespectState
	var userUsed int32
	var petUsed int32
	err := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `select coalesce((select received from player_respect_totals where player_id=$1),0),coalesce((select count(*) from player_respect_grants where actor_player_id=$1 and grant_date=$2),0),coalesce((select count(*) from pet_respects where actor_player_id=$1 and respected_on=$2),0)`, playerID, date.Format("2006-01-02")).Scan(&state.Received, &userUsed, &petUsed)
	if err != nil {
		return playerprofile.RespectState{}, fmt.Errorf("read player respect state: %w", err)
	}
	state.UserRemaining = remaining(userLimit, userUsed)
	state.PetRemaining = remaining(petLimit, petUsed)
	return state, nil
}

// GrantRespect serializes and applies one daily user respect.
func (repository *Repository) GrantRespect(ctx context.Context, actorID int64, targetID int64, date time.Time, limit int, unlimited bool) (result playerprofile.RespectResult, err error) {
	err = postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		var grantErr error
		result, grantErr = grantRespect(txCtx, postgres.ExecutorFor(txCtx, repository.pool), actorID, targetID, date.Format("2006-01-02"), limit, unlimited)
		return grantErr
	})
	return result, err
}

// grantRespect applies one serialized respect grant through the active transaction.
func grantRespect(ctx context.Context, executor postgres.Executor, actorID int64, targetID int64, dateValue string, limit int, unlimited bool) (result playerprofile.RespectResult, err error) {
	var lockedActorID int64
	if lockErr := executor.QueryRow(ctx, `select id from players where id=$1 and deleted_at is null for update`, actorID).Scan(&lockedActorID); lockErr != nil {
		return result, fmt.Errorf("lock respect actor: %w", lockErr)
	}
	var used int32
	if countErr := executor.QueryRow(ctx, `select count(*) from player_respect_grants where actor_player_id=$1 and grant_date=$2`, actorID, dateValue).Scan(&used); countErr != nil {
		return result, fmt.Errorf("count daily respect grants: %w", countErr)
	}
	result.Remaining = remaining(limit, used)
	if !unlimited && result.Remaining == 0 {
		return result, nil
	}
	command, insertErr := executor.Exec(ctx, `insert into player_respect_grants(actor_player_id,target_player_id,grant_date,source) values($1,$2,$3,'user') on conflict do nothing`, actorID, targetID, dateValue)
	if insertErr != nil {
		return result, fmt.Errorf("insert daily respect grant: %w", insertErr)
	}
	if command.RowsAffected() == 0 {
		result.Duplicate = true
		return result, nil
	}
	sourceKey := fmt.Sprintf("user:%d:%d:%s", actorID, targetID, dateValue)
	command, insertErr = executor.Exec(ctx, `insert into player_respect_ledger(source_key,player_id,amount,source) values($1,$2,1,'user') on conflict do nothing`, sourceKey, targetID)
	if insertErr != nil {
		return result, fmt.Errorf("insert respect ledger entry: %w", insertErr)
	}
	if command.RowsAffected() == 0 {
		return result, fmt.Errorf("insert respect ledger entry: source key %q already exists", sourceKey)
	}
	if scanErr := executor.QueryRow(ctx, `insert into player_respect_totals(player_id,received) values($1,1) on conflict(player_id) do update set received=player_respect_totals.received+1,updated_at=now() returning received`, targetID).Scan(&result.TotalReceived); scanErr != nil {
		return result, fmt.Errorf("update received respect total: %w", scanErr)
	}
	result.Applied = true
	if unlimited {
		result.Remaining = int32(limit)
	} else {
		result.Remaining--
	}
	return result, nil
}

// remaining computes a bounded daily allowance.
func remaining(limit int, used int32) int32 {
	value := int32(limit) - used
	if value < 0 {
		return 0
	}
	return value
}
