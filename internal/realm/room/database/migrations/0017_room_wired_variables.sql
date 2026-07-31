--liquibase formatted sql

--changeset pixels:room-wired-variables
create table room_wired_variables (
    id bigserial primary key,
    room_id bigint not null references rooms(id) on delete cascade,
    scope smallint not null,
    scope_id bigint not null,
    name text not null,
    int_value bigint not null default 0,
    string_value text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint room_wired_variables_scope_chk check (scope between 1 and 4),
    constraint room_wired_variables_scope_id_chk check (scope_id > 0),
    constraint room_wired_variables_name_chk check (length(name) between 1 and 64),
    constraint room_wired_variables_string_chk check (length(string_value) <= 2048),
    constraint room_wired_variables_unique unique(room_id,scope,scope_id,name)
);

create index room_wired_variables_scope_idx
    on room_wired_variables(room_id,scope,scope_id);

--rollback drop table if exists room_wired_variables;
