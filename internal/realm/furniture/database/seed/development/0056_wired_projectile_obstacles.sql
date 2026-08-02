--liquibase formatted sql

--changeset pixels:furniture-seed-wired-projectile-obstacles-0056 context:development
with fixture(id,definition_name,x,y,z,rotation,extra_data) as (values
 (1030700,'wf_xtra_projectile',5,3,0,0,'0'),
 (1030701,'wf_trg_says_something',5,3,0,0,'0'),
 (1030702,'table_plasto_4leg',5,7,0,0,'0'),
 (1030703,'wf_xtra_projectile',6,3,0,0,'0'),
 (1030704,'wf_trg_says_something',6,3,0,0,'0'),
 (1030705,'table_plasto_4leg',10,9,0,0,'0'),
 (1030706,'table_plasto_4leg',7,5,0,0,'0')
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

update furniture_items item
set owner_player_id=room.owner_player_id,x=5,y=5,z=0,rotation=0,
 deleted_at=null,updated_at=now()
from rooms room
where item.id=1020062 and room.id=207 and item.room_id=room.id;

with desired(item_id,int_params,string_param,selection_mode,delay_pulses) as (values
 (1030700,'[2,5,1]','',1,0),
 (1030701,'[]','user',0,0),
 (1030703,'[2,5,1]','',1,0),
 (1030704,'[]','border',0,0)
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
 (1030700,1030702,0,'0',5,7,0,0),
 (1030703,1030705,0,'0',10,9,0,0)
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

select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where room_id=207 and id between 1030700 and 1030706;
