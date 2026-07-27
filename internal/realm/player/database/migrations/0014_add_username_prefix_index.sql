--liquibase formatted sql

--changeset pixels:pixels-player-0014-add-username-prefix-index
--validCheckSum: 9:6159bd470d8fbd72a18152d4ca030387
create index if not exists players_username_prefix_idx
on players (lower(username) text_pattern_ops)
where deleted_at is null;

create index players_created_idx
on players (created_at desc, id desc)
where deleted_at is null;

--rollback drop index if exists players_created_idx; drop index if exists players_username_prefix_idx;
