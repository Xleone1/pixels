--liquibase formatted sql

--changeset pixels:room-0019-branding-drop-audit
drop table room_branding_audit;

--rollback create table room_branding_audit (
--rollback     id bigserial primary key,
--rollback     branding_id bigint not null references room_brandings(id),
--rollback     room_id bigint not null references rooms(id),
--rollback     furniture_item_id bigint not null references furniture_items(id),
--rollback     actor_player_id bigint not null references players(id),
--rollback     action text not null,
--rollback     reason text not null,
--rollback     before_state jsonb not null,
--rollback     after_state jsonb not null,
--rollback     created_at timestamptz not null default now(),
--rollback     constraint room_branding_audit_action_chk check (action in ('created', 'updated', 'disabled'))
--rollback );
--rollback create index room_branding_audit_room_idx
--rollback on room_branding_audit (room_id, id desc);
