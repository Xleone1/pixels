// Package insight reads aggregate room observability from PostgreSQL.
package insight

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roominsight "github.com/niflaot/pixels/internal/realm/room/record/insight"
	"github.com/niflaot/pixels/pkg/postgres"
)

const statsSQL = `
select room.id,room.max_users,
  (select count(*) from furniture_items item where item.room_id=room.id and item.deleted_at is null),
  (select count(*) from furniture_items item join furniture_definitions definition on definition.id=item.definition_id where item.room_id=room.id and item.deleted_at is null and definition.kind='floor'),
  (select count(*) from furniture_items item join furniture_definitions definition on definition.id=item.definition_id where item.room_id=room.id and item.deleted_at is null and definition.kind='wall'),
  (select count(*) from furniture_items item join furniture_definitions definition on definition.id=item.definition_id where item.room_id=room.id and item.deleted_at is null and definition.interaction_type like 'wf_%'),
  (select count(*) from bots where room_id=room.id),
  (select count(*) from pets where room_id=room.id and deleted_at is null),
  (select count(*) from room_brandings where room_id=room.id and enabled)
from rooms room where room.id=$1 and room.deleted_at is null`

// Repository reads durable room aggregate counters.
type Repository struct {
	// pool executes PostgreSQL queries.
	pool *postgres.Pool
}

// New creates a room insight repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Stats reads durable counters for one room.
func (repository *Repository) Stats(ctx context.Context, roomID int64) (roominsight.Stats, bool, error) {
	var stats roominsight.Stats
	err := repository.pool.QueryRow(ctx, statsSQL, roomID).Scan(
		&stats.RoomID, &stats.MaxUsers, &stats.Furniture, &stats.FloorFurniture,
		&stats.WallFurniture, &stats.WiredFurniture, &stats.Bots, &stats.Pets, &stats.Branding,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return stats, false, nil
	}
	if err != nil {
		return stats, false, fmt.Errorf("read room insight: %w", err)
	}
	return stats, true, nil
}

var storeAssertion roominsight.Store = (*Repository)(nil)
