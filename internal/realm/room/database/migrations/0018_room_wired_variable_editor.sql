--liquibase formatted sql

--changeset pixels:room-0018-wired-variable-editor
alter table room_wired_variables
    add column updated_by_player_id bigint null;

create index room_wired_variables_editor_idx
    on room_wired_variables(updated_by_player_id)
    where updated_by_player_id is not null;

--rollback drop index if exists room_wired_variables_editor_idx;
--rollback alter table room_wired_variables drop column if exists updated_by_player_id;
