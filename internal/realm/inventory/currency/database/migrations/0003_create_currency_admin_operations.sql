--liquibase formatted sql

--changeset pixels:pixels-currency-0003-admin-operations
create table currency_admin_operations (
    operation_key uuid primary key,
    request_hash text not null,
    player_id bigint not null references players(id),
    currency_type integer not null,
    balance_after bigint not null,
    delta bigint not null,
    created_at timestamptz not null default now()
);

create index currency_admin_operations_created_idx
    on currency_admin_operations(created_at desc);

--rollback drop table if exists currency_admin_operations;
