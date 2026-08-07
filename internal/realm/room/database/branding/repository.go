// Package branding persists room branding configuration.
package branding

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roombranding "github.com/niflaot/pixels/internal/realm/room/record/branding"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	listSQL             = `select id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version from room_brandings where room_id=$1 order by enabled desc,id`
	compatibleSQL       = `select item.id,definition.id,definition.name,definition.public_name,case when definition.interaction_type='furniture_bb' then 'billboard' when definition.interaction_type='furniture_bg' then 'background' else definition.metadata->>'branding_kind' end,branding.id is not null from furniture_items item join furniture_definitions definition on definition.id=item.definition_id left join room_brandings branding on branding.furniture_item_id=item.id where item.room_id=$1 and item.deleted_at is null and definition.deleted_at is null and item.x is not null and item.y is not null and item.z is not null and (definition.interaction_type in ('furniture_bg','furniture_bb','room_branding') or definition.metadata->>'branding_kind' in ('background','billboard')) order by item.id`
	lockItemSQL         = `select item.owner_player_id,item.x,item.y,item.z::float8,item.rotation,definition.id,definition.sprite_id,definition.interaction_type,definition.metadata->>'branding_kind' from furniture_items item join furniture_definitions definition on definition.id=item.definition_id where item.id=$1 and item.room_id=$2 and item.deleted_at is null and definition.deleted_at is null and item.x is not null and item.y is not null and item.z is not null for update of item`
	findByItemSQL       = `select id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version from room_brandings where furniture_item_id=$1`
	findByIDSQL         = `select id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version from room_brandings where id=$1 and room_id=$2`
	insertSQL           = `insert into room_brandings(room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,created_by_player_id,updated_by_player_id) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11) returning id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version`
	updateSQL           = `update room_brandings set kind=$3,asset_ref=$4,image_url=$5,click_url=$6,state=$7,offset_x=$8,offset_y=$9,offset_z=$10,enabled=true,updated_by_player_id=$11,updated_at=now(),version=version+1 where furniture_item_id=$1 and version=$2 returning id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version`
	disableSQL          = `update room_brandings set enabled=false,updated_by_player_id=$4,updated_at=now(),version=version+1 where id=$1 and room_id=$2 and version=$3 and enabled returning id,room_id,furniture_item_id,kind,asset_ref,image_url,click_url,state,offset_x,offset_y,offset_z,enabled,created_by_player_id,updated_by_player_id,created_at,updated_at,version`
	updateItemSQL       = `update furniture_items set extra_data=$2,updated_at=now(),version=version+1 where id=$1`
	updateDefinitionSQL = `update furniture_definitions set interaction_type=$2,updated_at=now(),version=version+1 where id=$1 and (interaction_type='room_branding' or metadata->>'branding_kind' in ('background','billboard'))`
)

// Repository persists room branding records.
type Repository struct {
	// pool owns transactional database access.
	pool *postgres.Pool
}

// New creates a room branding repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// List lists branding configurations for one room.
func (repository *Repository) List(ctx context.Context, roomID int64) ([]roombranding.Config, error) {
	rows, err := repository.pool.Query(ctx, listSQL, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room branding: %w", err)
	}
	defer rows.Close()
	items := make([]roombranding.Config, 0)
	for rows.Next() {
		item, scanErr := scanConfig(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListCompatible lists branding-capable furniture placed in one room.
func (repository *Repository) ListCompatible(ctx context.Context, roomID int64) ([]roombranding.CompatibleItem, error) {
	rows, err := repository.pool.Query(ctx, compatibleSQL, roomID)
	if err != nil {
		return nil, fmt.Errorf("list compatible branding furniture: %w", err)
	}
	defer rows.Close()
	items := make([]roombranding.CompatibleItem, 0)
	for rows.Next() {
		var item roombranding.CompatibleItem
		if err = rows.Scan(&item.ItemID, &item.DefinitionID, &item.Name, &item.PublicName, &item.Kind, &item.Configured); err != nil {
			return nil, fmt.Errorf("scan compatible branding furniture: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Upsert creates or updates branding and furniture state atomically.
func (repository *Repository) Upsert(ctx context.Context, mutation roombranding.Mutation, extraData string) (roombranding.Projection, error) {
	var result roombranding.Projection
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		item, err := lockItem(txCtx, executor, mutation.RoomID, mutation.FurnitureItemID, mutation.Kind)
		if err != nil {
			return err
		}
		_, found, err := findConfig(txCtx, executor, findByItemSQL, mutation.FurnitureItemID)
		if err != nil {
			return err
		}
		var after roombranding.Config
		if found {
			after, err = scanConfig(executor.QueryRow(txCtx, updateSQL, mutation.FurnitureItemID, mutation.ExpectedVersion, mutation.Kind, mutation.AssetRef, mutation.ImageURL, mutation.ClickURL, mutation.State, mutation.OffsetX, mutation.OffsetY, mutation.OffsetZ, mutation.ActorPlayerID))
		} else if mutation.ExpectedVersion == 0 {
			after, err = scanConfig(executor.QueryRow(txCtx, insertSQL, mutation.RoomID, mutation.FurnitureItemID, mutation.Kind, mutation.AssetRef, mutation.ImageURL, mutation.ClickURL, mutation.State, mutation.OffsetX, mutation.OffsetY, mutation.OffsetZ, mutation.ActorPlayerID))
		} else {
			return roombranding.ErrConflict
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return roombranding.ErrConflict
		}
		if err != nil {
			return fmt.Errorf("persist room branding: %w", err)
		}
		if _, err = executor.Exec(txCtx, updateItemSQL, mutation.FurnitureItemID, extraData); err != nil {
			return fmt.Errorf("update branding furniture state: %w", err)
		}
		interactionType := "furniture_bg"
		if mutation.Kind == roombranding.KindBillboard {
			interactionType = "furniture_bb"
		}
		if _, err = executor.Exec(txCtx, updateDefinitionSQL, item.definitionID, interactionType); err != nil {
			return fmt.Errorf("update branding furniture definition: %w", err)
		}
		result = item.projection(after, extraData, interactionType)
		return nil
	})
	return result, err
}

// Disable disables branding and clears projected furniture state atomically.
func (repository *Repository) Disable(ctx context.Context, roomID int64, brandingID int64, expectedVersion int64, actorPlayerID int64, extraData string) (roombranding.Projection, error) {
	var result roombranding.Projection
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		before, found, err := findConfig(txCtx, executor, findByIDSQL, brandingID, roomID)
		if err != nil || !found {
			return roombranding.ErrNotFound
		}
		item, err := lockItem(txCtx, executor, roomID, before.FurnitureItemID, before.Kind)
		if err != nil {
			return err
		}
		after, err := scanConfig(executor.QueryRow(txCtx, disableSQL, brandingID, roomID, expectedVersion, actorPlayerID))
		if errors.Is(err, pgx.ErrNoRows) {
			return roombranding.ErrConflict
		}
		if err != nil {
			return fmt.Errorf("disable room branding: %w", err)
		}
		if _, err = executor.Exec(txCtx, updateItemSQL, before.FurnitureItemID, extraData); err != nil {
			return fmt.Errorf("clear branding furniture state: %w", err)
		}
		result = item.projection(after, extraData, item.interactionType)
		return nil
	})
	return result, err
}

// persistedItem contains locked furniture projection state.
type persistedItem struct {
	ownerPlayerID   int64
	x               int
	y               int
	z               float64
	rotation        int
	definitionID    int64
	spriteID        int
	interactionType string
}

// projection builds a committed live projection.
func (item persistedItem) projection(config roombranding.Config, extraData string, interactionType string) roombranding.Projection {
	return roombranding.Projection{Config: config, SpriteID: item.spriteID, OwnerPlayerID: item.ownerPlayerID, X: item.x, Y: item.y, Z: item.z, Rotation: item.rotation, ExtraData: extraData, InteractionType: interactionType}
}

// lockItem locks and validates one branding-capable floor furniture item.
func lockItem(ctx context.Context, executor postgres.Executor, roomID int64, itemID int64, kind roombranding.Kind) (persistedItem, error) {
	var item persistedItem
	var metadataKind *string
	err := executor.QueryRow(ctx, lockItemSQL, itemID, roomID).Scan(&item.ownerPlayerID, &item.x, &item.y, &item.z, &item.rotation, &item.definitionID, &item.spriteID, &item.interactionType, &metadataKind)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, roombranding.ErrNotFound
	}
	if err != nil {
		return item, fmt.Errorf("lock branding furniture: %w", err)
	}
	compatible := item.interactionType == "room_branding" || item.interactionType == "furniture_bg" && kind == roombranding.KindBackground || item.interactionType == "furniture_bb" && kind == roombranding.KindBillboard || metadataKind != nil && *metadataKind == string(kind)
	if !compatible {
		return item, roombranding.ErrIncompatible
	}
	return item, nil
}

// findConfig finds one branding record with a variable identifier shape.
func findConfig(ctx context.Context, executor postgres.Executor, query string, arguments ...any) (roombranding.Config, bool, error) {
	value, err := scanConfig(executor.QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return roombranding.Config{}, false, nil
	}
	return value, err == nil, err
}

// scanConfig scans one branding record.
func scanConfig(row pgx.Row) (value roombranding.Config, err error) {
	err = row.Scan(&value.ID, &value.RoomID, &value.FurnitureItemID, &value.Kind, &value.AssetRef, &value.ImageURL, &value.ClickURL, &value.State, &value.OffsetX, &value.OffsetY, &value.OffsetZ, &value.Enabled, &value.CreatedByPlayerID, &value.UpdatedByPlayerID, &value.CreatedAt, &value.UpdatedAt, &value.Version)
	return value, err
}

var storeAssertion roombranding.Store = (*Repository)(nil)
