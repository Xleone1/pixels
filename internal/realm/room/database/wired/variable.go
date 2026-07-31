package wired

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	wiredvariable "github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// LoadRoom loads every durable WIRED variable assignment for one room.
func (repository *Repository) LoadRoomVariables(ctx context.Context, roomID int64) ([]wiredvariable.Value, error) {
	rows, err := repository.pool.Query(ctx, `select room_id,scope,scope_id,name,int_value,string_value,created_at,updated_at from room_wired_variables where room_id=$1 order by scope,scope_id,name`, roomID)
	if err != nil {
		return nil, fmt.Errorf("load room WIRED variables: %w", err)
	}
	defer rows.Close()
	values := make([]wiredvariable.Value, 0)
	for rows.Next() {
		var value wiredvariable.Value
		if err := rows.Scan(&value.RoomID, &value.Scope, &value.ScopeID, &value.Name, &value.IntValue, &value.StringValue, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// FindVariable returns one exact durable WIRED variable assignment.
func (repository *Repository) FindVariable(
	ctx context.Context,
	roomID int64,
	scope wiredvariable.Scope,
	scopeID int64,
	name string,
) (wiredvariable.Value, bool, error) {
	value := wiredvariable.Value{
		RoomID: roomID, Scope: scope, ScopeID: scopeID, Name: name,
	}
	err := repository.pool.QueryRow(
		ctx,
		`select int_value,string_value,created_at,updated_at from room_wired_variables where room_id=$1 and scope=$2 and scope_id=$3 and name=$4`,
		roomID, scope, scopeID, name,
	).Scan(
		&value.IntValue, &value.StringValue,
		&value.CreatedAt, &value.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return wiredvariable.Value{}, false, nil
	}
	if err != nil {
		return wiredvariable.Value{}, false, fmt.Errorf(
			"find room WIRED variable: %w", err,
		)
	}
	return value, true, nil
}

// SetVariable creates or replaces one WIRED variable assignment.
func (repository *Repository) SetVariable(ctx context.Context, value wiredvariable.Value) (wiredvariable.Value, error) {
	err := repository.pool.QueryRow(ctx, `insert into room_wired_variables(room_id,scope,scope_id,name,int_value,string_value) values($1,$2,$3,$4,$5,$6) on conflict(room_id,scope,scope_id,name) do update set int_value=excluded.int_value,string_value=excluded.string_value,updated_at=now() returning created_at,updated_at`, value.RoomID, value.Scope, value.ScopeID, value.Name, value.IntValue, value.StringValue).Scan(&value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		return wiredvariable.Value{}, fmt.Errorf("set room WIRED variable: %w", err)
	}
	return value, nil
}

// DeleteVariable removes one exact WIRED variable assignment.
func (repository *Repository) DeleteVariable(ctx context.Context, roomID int64, scope wiredvariable.Scope, scopeID int64, name string) (bool, error) {
	var deleted int
	err := repository.pool.QueryRow(ctx, `delete from room_wired_variables where room_id=$1 and scope=$2 and scope_id=$3 and name=$4 returning 1`, roomID, scope, scopeID, name).Scan(&deleted)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return deleted == 1, err
}

// variableStore adapts the WIRED repository to the focused variable contract.
type variableStore struct {
	// repository persists variable rows.
	repository *Repository
}

// LoadRoom loads every assignment through the shared repository.
func (store variableStore) LoadRoom(ctx context.Context, roomID int64) ([]wiredvariable.Value, error) {
	return store.repository.LoadRoomVariables(ctx, roomID)
}

// Find returns one exact durable assignment.
func (store variableStore) Find(
	ctx context.Context,
	roomID int64,
	scope wiredvariable.Scope,
	scopeID int64,
	name string,
) (wiredvariable.Value, bool, error) {
	return store.repository.FindVariable(ctx, roomID, scope, scopeID, name)
}

// Set creates or replaces one assignment.
func (store variableStore) Set(ctx context.Context, value wiredvariable.Value) (wiredvariable.Value, error) {
	return store.repository.SetVariable(ctx, value)
}

// Delete removes one assignment.
func (store variableStore) Delete(ctx context.Context, roomID int64, scope wiredvariable.Scope, scopeID int64, name string) (bool, error) {
	return store.repository.DeleteVariable(ctx, roomID, scope, scopeID, name)
}

// VariableStore exposes the focused variable persistence contract.
func (repository *Repository) VariableStore() wiredvariable.Store {
	return variableStore{repository: repository}
}
