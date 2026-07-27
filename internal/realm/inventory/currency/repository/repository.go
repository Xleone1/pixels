package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

// txRunner runs currency work inside one transaction.
type txRunner func(context.Context, func(context.Context, postgres.Executor) error) error

// Repository reads and writes currency persistence records.
type Repository struct {
	// executor runs read-only PostgreSQL queries.
	executor postgres.Executor

	// withinTx runs atomic currency mutations.
	withinTx txRunner
}

// New creates a currency repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{
		executor: pool,
		withinTx: func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			if executor, ok := postgres.ScopedExecutor(ctx); ok {
				return work(ctx, executor)
			}
			return postgres.WithinTx(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
				return work(ctx, tx)
			})
		},
	}
}

// FindBalance finds one player currency balance.
func (repository *Repository) FindBalance(ctx context.Context, playerID int64, currencyType int32) (currencymodel.Balance, bool, error) {
	return findBalance(ctx, repository.executor, playerID, currencyType)
}

// ListBalances lists one player's stored currency balances.
func (repository *Repository) ListBalances(ctx context.Context, playerID int64) ([]currencymodel.Balance, error) {
	rows, err := repository.executor.Query(ctx, `
		select player_id, currency_type, amount, updated_at, version
		from player_currencies
		where player_id = $1
		order by currency_type`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d currency balances: %w", playerID, err)
	}
	defer rows.Close()

	balances := make([]currencymodel.Balance, 0)
	for rows.Next() {
		var balance currencymodel.Balance
		if err := rows.Scan(&balance.PlayerID, &balance.CurrencyType, &balance.Amount, &balance.UpdatedAt, &balance.Version); err != nil {
			return nil, fmt.Errorf("scan player %d currency balance: %w", playerID, err)
		}
		balances = append(balances, balance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player %d currency balances: %w", playerID, err)
	}

	return balances, nil
}

// Grant applies a signed delta and optional ledger entry atomically.
func (repository *Repository) Grant(ctx context.Context, mutation Mutation) (Result, error) {
	return repository.mutate(ctx, mutation, false)
}

// Set replaces a balance and writes an optional ledger entry atomically.
func (repository *Repository) Set(ctx context.Context, mutation Mutation) (Result, error) {
	return repository.mutate(ctx, mutation, true)
}

// mutate serializes and persists one currency mutation.
func (repository *Repository) mutate(ctx context.Context, mutation Mutation, absolute bool) (Result, error) {
	var result Result
	err := repository.withinTx(ctx, func(ctx context.Context, executor postgres.Executor) error {
		replayed, replayErr := replayOperation(ctx, executor, mutation, &result)
		if replayErr != nil || replayed {
			return replayErr
		}
		if err := lockBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType); err != nil {
			return err
		}

		current, found, err := findBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType)
		if err != nil {
			return err
		}
		previous := int64(0)
		if found {
			previous = current.Amount
		}

		if !absolute && mutation.Amount > 0 && previous > math.MaxInt64-mutation.Amount {
			return ErrBalanceOverflow
		}
		next := previous + mutation.Amount
		if absolute {
			next = mutation.Amount
		}
		if next < 0 {
			return ErrInsufficientBalance
		}

		result.Balance, err = upsertBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType, next)
		if err != nil {
			return err
		}
		result.Delta = next - previous
		if mutation.Ledger {
			if err = insertLedger(ctx, executor, ledgerEntry(mutation, result.Delta, next)); err != nil {
				return err
			}
		}
		return insertOperation(ctx, executor, mutation, result)
	})
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

// replayOperation locks and resolves one optional idempotency operation.
func replayOperation(ctx context.Context, executor postgres.Executor, mutation Mutation, result *Result) (bool, error) {
	if mutation.OperationKey == "" {
		return false, nil
	}
	if _, err := executor.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended($1, 0))`, mutation.OperationKey); err != nil {
		return false, fmt.Errorf("lock currency operation: %w", err)
	}
	var requestHash string
	err := executor.QueryRow(ctx, `
		select request_hash, balance_after, delta
		from currency_admin_operations
		where operation_key = $1`, mutation.OperationKey).Scan(
		&requestHash,
		&result.Balance.Amount,
		&result.Delta,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read currency operation: %w", err)
	}
	if requestHash != mutation.RequestHash {
		return false, ErrIdempotencyConflict
	}
	result.Balance.PlayerID = mutation.PlayerID
	result.Balance.CurrencyType = mutation.CurrencyType
	result.Replayed = true
	return true, nil
}

// insertOperation persists one optional replay result.
func insertOperation(ctx context.Context, executor postgres.Executor, mutation Mutation, result Result) error {
	if mutation.OperationKey == "" {
		return nil
	}
	_, err := executor.Exec(ctx, `
		insert into currency_admin_operations(
			operation_key, request_hash, player_id, currency_type, balance_after, delta
		) values ($1, $2, $3, $4, $5, $6)`,
		mutation.OperationKey,
		mutation.RequestHash,
		mutation.PlayerID,
		mutation.CurrencyType,
		result.Balance.Amount,
		result.Delta,
	)
	if err != nil {
		return fmt.Errorf("insert currency operation: %w", err)
	}
	return nil
}
