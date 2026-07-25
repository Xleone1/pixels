package profile

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// respectExecutor records the ordered persistence operations for one grant.
type respectExecutor struct {
	// queries stores issued row queries.
	queries []string
	// rows stores queued row results.
	rows [][]any
	// commands stores queued command results.
	commands []pgconn.CommandTag
}

// Exec records one mutation and returns its queued command result.
func (executor *respectExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	if len(executor.commands) == 0 {
		return pgconn.CommandTag{}, errors.New("unexpected exec")
	}
	command := executor.commands[0]
	executor.commands = executor.commands[1:]
	return command, nil
}

// Query rejects unused multi-row queries.
func (executor *respectExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

// QueryRow records one query and returns its queued row.
func (executor *respectExecutor) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	executor.queries = append(executor.queries, query)
	if len(executor.rows) == 0 {
		return respectRow{err: errors.New("unexpected query row")}
	}
	row := executor.rows[0]
	executor.rows = executor.rows[1:]
	return respectRow{values: row}
}

// respectRow provides one deterministic pgx row.
type respectRow struct {
	// values stores scan values.
	values []any
	// err stores one injected scan failure.
	err error
}

// Scan copies queued scalar values into supported destinations.
func (row respectRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	for index, destination := range destinations {
		switch target := destination.(type) {
		case *int64:
			*target = row.values[index].(int64)
		case *int32:
			*target = row.values[index].(int32)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}

// TestGrantRespectSerializesOnActor verifies transparent row locking and committed counters.
func TestGrantRespectSerializesOnActor(t *testing.T) {
	executor := &respectExecutor{
		rows:     [][]any{{int64(1)}, {int32(0)}, {int32(9)}},
		commands: []pgconn.CommandTag{pgconn.NewCommandTag("INSERT 0 1"), pgconn.NewCommandTag("INSERT 0 1")},
	}
	result, err := grantRespect(context.Background(), executor, 1, 6, "2026-07-25", 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || result.TotalReceived != 9 || result.Remaining != 2 {
		t.Fatalf("result=%+v", result)
	}
	if len(executor.queries) != 3 || !strings.Contains(executor.queries[0], "for update") {
		t.Fatalf("queries=%q", executor.queries)
	}
}

// TestGrantRespectRejectsDuplicateLedger verifies inconsistent idempotency state rolls back.
func TestGrantRespectRejectsDuplicateLedger(t *testing.T) {
	executor := &respectExecutor{
		rows:     [][]any{{int64(1)}, {int32(0)}},
		commands: []pgconn.CommandTag{pgconn.NewCommandTag("INSERT 0 1"), pgconn.NewCommandTag("INSERT 0 0")},
	}
	result, err := grantRespect(context.Background(), executor, 1, 6, "2026-07-25", 3, false)
	if err == nil || !strings.Contains(err.Error(), "source key") || result.Applied {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
