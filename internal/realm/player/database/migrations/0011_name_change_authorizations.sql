--liquibase formatted sql

--changeset pixels:pixels-player-0011-name-change-authorizations
create table player_name_change_authorizations (
    id bigserial primary key,
    player_id bigint not null references players(id),
    allowed boolean not null,
    actor_player_id bigint not null references players(id),
    reason varchar(500) not null,
    changed_at timestamptz not null default now()
);

create index player_name_change_authorizations_player_idx
    on player_name_change_authorizations(player_id, changed_at desc);

--rollback drop index if exists player_name_change_authorizations_player_idx; drop table if exists player_name_change_authorizations;
