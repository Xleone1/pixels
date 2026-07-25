--liquibase formatted sql

--changeset pixels:pixels-player-seed-test-0003-sync-player-identity
select setval(
    pg_get_serial_sequence('players', 'id'),
    coalesce((select max(id) from players), 1),
    exists(select 1 from players)
);

--rollback select setval(pg_get_serial_sequence('players', 'id'), coalesce((select max(id) from players), 1), exists(select 1 from players));
