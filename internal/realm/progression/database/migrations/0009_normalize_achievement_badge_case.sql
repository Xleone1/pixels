--liquibase formatted sql

--changeset pixels:pixels-progression-0009-normalize-achievement-badge-case
with canonical_badges as (
    select
        badge.player_id,
        badge.code as current_code,
        'ACH_' || definition.name || level.level::text as canonical_code
    from player_badges badge
    join achievement_definitions definition
      on lower(badge.code) like lower('ACH_' || definition.name) || '%'
    join achievement_levels level
      on lower(badge.code) = lower('ACH_' || definition.name || level.level::text)
    where badge.code <> 'ACH_' || definition.name || level.level::text
),
conflict_free as (
    select canonical.*
    from canonical_badges canonical
    where not exists (
        select 1
        from player_badges existing
        where existing.player_id = canonical.player_id
          and existing.code = canonical.canonical_code
    )
)
update player_badges badge
set code = canonical.canonical_code
from conflict_free canonical
where badge.player_id = canonical.player_id
  and badge.code = canonical.current_code;

--rollback select 1;
