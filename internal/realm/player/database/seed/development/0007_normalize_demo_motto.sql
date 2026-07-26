--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0007-normalize-demo-motto
update player_profiles
set motto = 'Welcome to Pixels.'
where player_id = 1
  and motto = 'Welcome to Pixels - make yourself at home.';

--rollback update player_profiles set motto = 'Welcome to Pixels - make yourself at home.' where player_id = 1 and motto = 'Welcome to Pixels.';
