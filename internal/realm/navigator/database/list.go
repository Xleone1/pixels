package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

const (
	// addFavoriteSQL inserts a player favorite.
	addFavoriteSQL = `insert into room_favorites(player_id,room_id) select $1,r.id from rooms r where r.id=$2 and r.deleted_at is null and not r.is_bundle_template and ($4 or (select count(*) from room_favorites where player_id=$1)<$3) on conflict do nothing`
	// removeFavoriteSQL deletes a player favorite.
	removeFavoriteSQL = `delete from room_favorites where player_id = $1 and room_id = $2`
	// listFavoriteRoomIDsSQL reads active favorite room identifiers through one bounded join.
	listFavoriteRoomIDsSQL = `select favorites.room_id from room_favorites favorites join rooms on rooms.id=favorites.room_id where favorites.player_id=$1 and rooms.deleted_at is null and not rooms.is_bundle_template order by favorites.created_at desc`
	// liftedRoomColumns contains the shared lifted room select list.
	liftedRoomColumns = `id, room_id, area_id, image, asset_ref, caption, order_num, starts_at, ends_at, created_by_player_id, updated_by_player_id, created_at, updated_at, deleted_at, version`
	// listLiftedRoomsSQL reads active lifted rooms.
	listLiftedRoomsSQL = `select ` + liftedRoomColumns + ` from navigator_lifted_rooms where deleted_at is null and (starts_at is null or starts_at <= now()) and (ends_at is null or ends_at > now()) order by order_num asc, id asc`
	// findActiveLiftedRoomSQL locks one room's active Navigator media.
	findActiveLiftedRoomSQL = `select ` + liftedRoomColumns + ` from navigator_lifted_rooms where room_id=$1 and deleted_at is null order by updated_at desc,id desc limit 1 for update`
	// insertLiftedRoomSQL creates one Navigator media row.
	insertLiftedRoomSQL = `insert into navigator_lifted_rooms(room_id,area_id,image,asset_ref,caption,order_num,starts_at,ends_at,created_by_player_id,updated_by_player_id) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$9) returning ` + liftedRoomColumns
	// updateLiftedRoomSQL replaces one active Navigator media row optimistically.
	updateLiftedRoomSQL = `update navigator_lifted_rooms set area_id=$2,image=$3,asset_ref=$4,caption=$5,order_num=$6,starts_at=$7,ends_at=$8,updated_by_player_id=$9,updated_at=now(),version=version+1 where id=$1 and version=$10 returning ` + liftedRoomColumns
	// disableLiftedRoomSQL soft deletes one active Navigator media row optimistically.
	disableLiftedRoomSQL = `update navigator_lifted_rooms set deleted_at=now(),updated_at=now(),updated_by_player_id=$3,version=version+1 where id=$1 and version=$2 returning ` + liftedRoomColumns
)

// AddFavorite adds a favorite room for a player.
func (repository *Repository) AddFavorite(ctx context.Context, playerID int64, roomID int64, limit int32, unlimited bool) error {
	result, err := repository.executor.Exec(ctx, addFavoriteSQL, playerID, roomID, limit, unlimited)
	if err != nil {
		return fmt.Errorf("add room favorite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return navmodel.ErrFavoriteUnavailable
	}
	return nil
}

// RemoveFavorite removes a favorite room for a player.
func (repository *Repository) RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error {
	if _, err := repository.executor.Exec(ctx, removeFavoriteSQL, playerID, roomID); err != nil {
		return fmt.Errorf("remove room favorite: %w", err)
	}
	return nil
}

// ListFavoriteRoomIDs lists favorite room identifiers for a player.
func (repository *Repository) ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := repository.executor.Query(ctx, listFavoriteRoomIDsSQL, playerID)
	if err != nil {
		return nil, fmt.Errorf("list room favorites: %w", err)
	}
	defer rows.Close()
	return scanFavoriteRoomIDs(rows)
}

// ListLiftedRooms lists currently active lifted rooms.
func (repository *Repository) ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error) {
	rows, err := repository.executor.Query(ctx, listLiftedRoomsSQL)
	if err != nil {
		return nil, fmt.Errorf("list lifted rooms: %w", err)
	}
	defer rows.Close()
	return scanLiftedRooms(rows)
}

// UpsertLiftedRoom creates or updates one room's active Navigator media.
func (repository *Repository) UpsertLiftedRoom(ctx context.Context, mutation navmodel.LiftedRoomMutation) (navmodel.LiftedRoom, error) {
	active, found, err := repository.findActiveLiftedRoom(ctx, mutation.RoomID)
	if err != nil {
		return navmodel.LiftedRoom{}, err
	}
	if !found {
		if mutation.ExpectedVersion != 0 {
			return navmodel.LiftedRoom{}, navmodel.ErrLiftedRoomConflict
		}
		value, scanErr := scanLiftedRoom(repository.executor.QueryRow(ctx, insertLiftedRoomSQL, mutation.RoomID, mutation.AreaID, mutation.Image, mutation.AssetRef, mutation.Caption, mutation.Order, mutation.StartsAt, mutation.EndsAt, mutation.ActorPlayerID))
		if scanErr != nil {
			return navmodel.LiftedRoom{}, fmt.Errorf("create lifted room media: %w", scanErr)
		}
		return value, nil
	}
	value, err := scanLiftedRoom(repository.executor.QueryRow(ctx, updateLiftedRoomSQL, active.ID, mutation.AreaID, mutation.Image, mutation.AssetRef, mutation.Caption, mutation.Order, mutation.StartsAt, mutation.EndsAt, mutation.ActorPlayerID, mutation.ExpectedVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return navmodel.LiftedRoom{}, navmodel.ErrLiftedRoomConflict
	}
	if err != nil {
		return navmodel.LiftedRoom{}, fmt.Errorf("update lifted room media: %w", err)
	}
	return value, nil
}

// DisableLiftedRoom soft deletes one active Navigator media row.
func (repository *Repository) DisableLiftedRoom(ctx context.Context, roomID int64, expectedVersion int64, actorPlayerID int64) (navmodel.LiftedRoom, error) {
	active, found, err := repository.findActiveLiftedRoom(ctx, roomID)
	if err != nil {
		return navmodel.LiftedRoom{}, err
	}
	if !found {
		return navmodel.LiftedRoom{}, navmodel.ErrLiftedRoomNotFound
	}
	value, err := scanLiftedRoom(repository.executor.QueryRow(ctx, disableLiftedRoomSQL, active.ID, expectedVersion, actorPlayerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return navmodel.LiftedRoom{}, navmodel.ErrLiftedRoomConflict
	}
	if err != nil {
		return navmodel.LiftedRoom{}, fmt.Errorf("disable lifted room media: %w", err)
	}
	return value, nil
}

// findActiveLiftedRoom reads one room's latest active Navigator media.
func (repository *Repository) findActiveLiftedRoom(ctx context.Context, roomID int64) (navmodel.LiftedRoom, bool, error) {
	value, err := scanLiftedRoom(repository.executor.QueryRow(ctx, findActiveLiftedRoomSQL, roomID))
	if errors.Is(err, pgx.ErrNoRows) {
		return navmodel.LiftedRoom{}, false, nil
	}
	if err != nil {
		return navmodel.LiftedRoom{}, false, fmt.Errorf("find active lifted room media: %w", err)
	}
	return value, true, nil
}
