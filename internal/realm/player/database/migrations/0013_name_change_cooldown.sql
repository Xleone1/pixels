--liquibase formatted sql

--changeset pixels:pixels-player-0013-name-change-cooldown
alter table player_profiles
    add column last_name_change_at timestamptz null;

update player_profiles profile
set last_name_change_at = latest.changed_at
from (
    select distinct on (player_id) player_id, changed_at
    from player_name_changes
    order by player_id, changed_at desc, id desc
) latest
where latest.player_id = profile.player_id;

drop table player_name_change_authorizations;

alter table player_profiles
    drop column allow_name_change;

--rollback alter table player_profiles add column allow_name_change boolean not null default false;
--rollback create table player_name_change_authorizations (id bigserial primary key, player_id bigint not null references players(id), allowed boolean not null, actor_player_id bigint not null references players(id), reason varchar(500) not null, changed_at timestamptz not null default now());
--rollback create index player_name_change_authorizations_player_idx on player_name_change_authorizations(player_id, changed_at desc);
--rollback alter table player_profiles drop column last_name_change_at;
