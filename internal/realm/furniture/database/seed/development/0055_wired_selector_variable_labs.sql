--liquibase formatted sql

--changeset pixels:furniture-seed-wired-selector-variable-labs-0055 context:development
with fixtures as (
 select item.id,
  case when definition.interaction_type like 'wf_slc_%' then 200 else 201 end room_id,
  row_number() over(
   partition by definition.interaction_type like 'wf_slc_%'
   order by definition.interaction_type,item.id
  ) - 1 ordinal
 from furniture_items item
 join furniture_definitions definition on definition.id=item.definition_id
 where item.id between 1010000 and 1010053
  and definition.deleted_at is null
  and (
   definition.interaction_type like 'wf_slc_%'
   or definition.interaction_type like 'wf_var_%'
   or definition.interaction_type='wf_trg_var_changed'
   or definition.interaction_type like 'wf_act_%var%'
   or definition.interaction_type like 'wf_cnd_%var%'
   or definition.interaction_type like 'wf_xtra_filter_%_by_var'
  )
), placed as (
 select fixtures.id,fixtures.room_id,
  2 + fixtures.ordinal % 10 x,2 + fixtures.ordinal / 10 y
 from fixtures
)
update furniture_items item
set owner_player_id=room.owner_player_id,room_id=room.id,
 x=placed.x,y=placed.y,z=0,rotation=0,deleted_at=null,updated_at=now()
from placed
join rooms room on room.id=placed.room_id
where item.id=placed.id;

update furniture_items item
set owner_player_id=room.owner_player_id,deleted_at=null,updated_at=now()
from rooms room
where item.room_id=room.id and room.id in (200,201)
 and item.id in (1020900,1020901);

--rollback update furniture_items item set owner_player_id=1,room_id=200 + definition.ordinal / 9,x=4 + definition.ordinal % 9 % 8,y=2 + definition.ordinal % 9 / 8,z=0,rotation=0 from (select id,row_number() over(order by id)-1 ordinal from furniture_definitions where metadata->>'source'='polaris-wired') definition where item.id=1010000 + definition.ordinal;
