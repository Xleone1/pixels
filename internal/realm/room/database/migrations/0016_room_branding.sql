--liquibase formatted sql

--changeset pixels:room-0016-branding
create table room_brandings (
    id bigserial primary key,
    room_id bigint not null references rooms(id),
    furniture_item_id bigint not null unique references furniture_items(id),
    kind text not null,
    asset_ref text not null default '',
    image_url text not null,
    click_url text not null default '',
    state smallint not null default 0,
    offset_x integer not null default 0,
    offset_y integer not null default 0,
    offset_z integer not null default 0,
    enabled boolean not null default true,
    created_by_player_id bigint not null references players(id),
    updated_by_player_id bigint not null references players(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint room_brandings_kind_chk check (kind in ('background', 'billboard')),
    constraint room_brandings_state_chk check (state between 0 and 255),
    constraint room_brandings_offset_x_chk check (offset_x between -4096 and 4096),
    constraint room_brandings_offset_y_chk check (offset_y between -4096 and 4096),
    constraint room_brandings_offset_z_chk check (offset_z between -4096 and 4096)
);

create index room_brandings_room_idx
on room_brandings (room_id, enabled, id)
where enabled;

create table room_branding_audit (
    id bigserial primary key,
    branding_id bigint not null references room_brandings(id),
    room_id bigint not null references rooms(id),
    furniture_item_id bigint not null references furniture_items(id),
    actor_player_id bigint not null references players(id),
    action text not null,
    reason text not null,
    before_state jsonb not null,
    after_state jsonb not null,
    created_at timestamptz not null default now(),
    constraint room_branding_audit_action_chk check (action in ('created', 'updated', 'disabled'))
);

create index room_branding_audit_room_idx
on room_branding_audit (room_id, id desc);

--rollback drop table room_branding_audit; drop table room_brandings;
