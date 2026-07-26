--liquibase formatted sql

--changeset pixels:pixels-player-0014-add-username-prefix-index
create index players_username_prefix_idx
on players (lower(username) text_pattern_ops)
where deleted_at is null;

create index players_created_idx
on players (created_at desc, id desc)
where deleted_at is null;

--rollback drop index if exists players_created_idx; drop index if exists players_username_prefix_idx;
