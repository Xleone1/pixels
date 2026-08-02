--liquibase formatted sql

--changeset pixels:furniture-seed-wired-projectile-0054 context:development
with desired(
 id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,
 allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,
 multiheight,custom_params,metadata) as (values
 (1000069,77201377,'wf_xtra_projectile','WIRED Custom: Projectile','floor',1,1,0.65,true,true,false,false,true,'wf_xtra_projectile',4,'','',
  '{"source":"pixels-wired-custom","wired_support":"native","asset_classname":"wf_xtra_exec_in_order"}'::jsonb)
)
insert into furniture_definitions(
 id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,
 allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,
 multiheight,custom_params,metadata)
overriding system value
select desired.* from desired
where not exists (
 select 1 from furniture_definitions existing
 where existing.name=desired.name and existing.deleted_at is null
)
on conflict(id) do nothing;

update furniture_definitions
set sprite_id=77201377,interaction_type='wf_xtra_projectile',
 metadata=jsonb_set(
  metadata,'{asset_classname}',to_jsonb('wf_xtra_exec_in_order'::text),true),
 deleted_at=null,updated_at=now()
where name='wf_xtra_projectile';

with fixture(id,definition_name,x,y,z,rotation,extra_data) as (values
 (1010069,'wf_xtra_projectile',4,2,0,0,'0'),
 (1020062,'table_plasto_4leg',5,5,0,0,'0'),
 (1020063,'wf_trg_says_something',4,2,0,0,'0')
)
insert into furniture_items(
 id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select fixture.id,definition.id,room.owner_player_id,room.id,
 fixture.x,fixture.y,fixture.z,fixture.rotation,fixture.extra_data
from fixture
join furniture_definitions definition on definition.name=fixture.definition_name
 and definition.deleted_at is null
join rooms room on room.id=207 and room.name='WIRED QA Projectile'
on conflict(id) do nothing;

with desired(item_id,int_params,string_param,selection_mode,delay_pulses) as (values
 (1010069,'[2,3,2]','',1,0),
 (1020063,'[]','launch',0,0)
)
insert into room_wired_settings(
 item_id,int_params,string_param,selection_mode,delay_pulses)
select desired.item_id,desired.int_params::jsonb,desired.string_param,
 desired.selection_mode,desired.delay_pulses
from desired
join furniture_items item on item.id=desired.item_id
where item.room_id=207 and item.deleted_at is null
on conflict(item_id) do update set int_params=excluded.int_params,
 string_param=excluded.string_param,selection_mode=excluded.selection_mode,
 delay_pulses=excluded.delay_pulses,version=room_wired_settings.version+1,
 updated_at=now();

with desired(
 wired_item_id,selected_item_id,ordinal,
 snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation) as (values
 (1010069,1020062,0,'0',5,5,0,0)
)
insert into room_wired_selected_items(
 wired_item_id,selected_item_id,ordinal,
 snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
select desired.*
from desired
join furniture_items wired on wired.id=desired.wired_item_id
join furniture_items selected on selected.id=desired.selected_item_id
where wired.room_id=207 and selected.room_id=207
 and wired.deleted_at is null and selected.deleted_at is null
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,
 snapshot_state=excluded.snapshot_state,snapshot_x=excluded.snapshot_x,
 snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,
 snapshot_rotation=excluded.snapshot_rotation;

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where room_id=207 and id in (1010069,1020062,1020063); delete from furniture_definitions where id=1000069 and metadata->>'source'='pixels-wired-custom';
