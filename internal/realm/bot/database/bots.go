package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// Find returns one bot with ordered chat lines.
func (repository *Repository) Find(ctx context.Context, botID int64) (botrecord.Bot, bool, error) {
	query := `select ` + botColumns + botFrom + `where b.id=$1` + botGroup
	bot, err := scanBot(repository.executorFor(ctx).QueryRow(ctx, query, botID))
	if errors.Is(err, pgx.ErrNoRows) {
		return botrecord.Bot{}, false, nil
	}
	return bot, err == nil, err
}

// Inventory lists bots held by one player.
func (repository *Repository) Inventory(ctx context.Context, playerID int64) ([]botrecord.Bot, error) {
	query := `select ` + botColumns + botFrom + `where b.owner_player_id=$1 and b.room_id is null` + botGroup + `order by b.id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	return scanBots(rows)
}

// Room lists bots placed in one room.
func (repository *Repository) Room(ctx context.Context, roomID int64) ([]botrecord.Bot, error) {
	query := `select ` + botColumns + botFrom + `where b.room_id=$1` + botGroup + `order by b.id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	return scanBots(rows)
}

// CountInventory counts bots held by one player.
func (repository *Repository) CountInventory(ctx context.Context, playerID int64) (int, error) {
	var count int
	err := repository.executorFor(ctx).QueryRow(ctx, `select count(*) from bots where owner_player_id=$1 and room_id is null`, playerID).Scan(&count)
	return count, err
}

// Create inserts one inventory bot.
func (repository *Repository) Create(ctx context.Context, bot botrecord.Bot) (botrecord.Bot, error) {
	const query = `insert into bots(owner_player_id,behavior_type,name,motto,figure,gender,can_walk,dance_type,chat_auto,chat_random,chat_delay_seconds,bubble_style,effect_id)
values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) returning id`
	if err := repository.executorFor(ctx).QueryRow(ctx, query, bot.OwnerPlayerID, bot.BehaviorType, bot.Name, bot.Motto, bot.Figure, bot.Gender, bot.CanWalk, bot.DanceType, bot.ChatAuto, bot.ChatRandom, bot.ChatDelaySeconds, bot.BubbleStyle, bot.EffectID).Scan(&bot.ID); err != nil {
		return botrecord.Bot{}, err
	}
	created, found, err := repository.Find(ctx, bot.ID)
	if err != nil {
		return botrecord.Bot{}, err
	}
	if !found {
		return botrecord.Bot{}, botrecord.ErrBotNotFound
	}

	return created, nil
}

// Place moves an owned inventory bot into a room.
func (repository *Repository) Place(ctx context.Context, botID int64, ownerID int64, roomID int64, x int, y int, z float64, rotation int16) (botrecord.Bot, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update bots set room_id=$3,x=$4,y=$5,z=$6,rotation=$7,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null`, botID, ownerID, roomID, x, y, z, rotation)
	if err != nil || command.RowsAffected() == 0 {
		return botrecord.Bot{}, false, err
	}
	return repository.Find(ctx, botID)
}

// Pickup moves a placed bot to a receiving owner.
func (repository *Repository) Pickup(ctx context.Context, botID int64, roomID int64, receiverID int64) (botrecord.Bot, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update bots set owner_player_id=$3,room_id=null,x=null,y=null,z=null,rotation=null,updated_at=now(),version=version+1 where id=$1 and room_id=$2`, botID, roomID, receiverID)
	if err != nil || command.RowsAffected() == 0 {
		return botrecord.Bot{}, false, err
	}
	return repository.Find(ctx, botID)
}

// ForcePickup moves a bot back to its existing owner.
func (repository *Repository) ForcePickup(ctx context.Context, botID int64) (botrecord.Bot, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update bots set room_id=null,x=null,y=null,z=null,rotation=null,updated_at=now(),version=version+1 where id=$1 and room_id is not null`, botID)
	if err != nil || command.RowsAffected() == 0 {
		return botrecord.Bot{}, false, err
	}
	return repository.Find(ctx, botID)
}

// Delete permanently removes one owned inventory bot.
func (repository *Repository) Delete(ctx context.Context, botID int64, ownerID int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `delete from bots where id=$1 and owner_player_id=$2 and room_id is null`, botID, ownerID)
	return err == nil && command.RowsAffected() > 0, err
}

// SavePosition persists a placed bot's latest world position.
func (repository *Repository) SavePosition(ctx context.Context, botID int64, roomID int64, x int, y int, z float64, rotation int16) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `update bots set x=$3,y=$4,z=$5,rotation=$6,updated_at=now(),version=version+1 where id=$1 and room_id=$2`, botID, roomID, x, y, z, rotation)
	if err != nil {
		return fmt.Errorf("save bot position: %w", err)
	}
	return nil
}
