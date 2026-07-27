--liquibase formatted sql

--changeset pixels:pixels-navigator-0003-lifted-room-assets
ALTER TABLE navigator_lifted_rooms
    ADD COLUMN asset_ref text NOT NULL DEFAULT '',
    ADD COLUMN created_by_player_id bigint NULL REFERENCES players(id),
    ADD COLUMN updated_by_player_id bigint NULL REFERENCES players(id);

CREATE INDEX navigator_lifted_rooms_room_active_idx
ON navigator_lifted_rooms (room_id, updated_at DESC)
WHERE deleted_at IS NULL;

--rollback DROP INDEX IF EXISTS navigator_lifted_rooms_room_active_idx;
--rollback ALTER TABLE navigator_lifted_rooms DROP COLUMN IF EXISTS updated_by_player_id;
--rollback ALTER TABLE navigator_lifted_rooms DROP COLUMN IF EXISTS created_by_player_id;
--rollback ALTER TABLE navigator_lifted_rooms DROP COLUMN IF EXISTS asset_ref;
