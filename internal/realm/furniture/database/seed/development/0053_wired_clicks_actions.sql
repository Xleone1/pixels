--liquibase formatted sql

--changeset pixels:furniture-seed-wired-clicks-actions-0053 context:development
with desired(
 id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,
 allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,
 multiheight,custom_params,metadata) as (values
 (1000060,48973618,'wf_trg_user_clicks_furni','WIRED: User Clicks Furniture','floor',1,1,0.65,true,true,false,false,true,'wf_trg_user_clicks_furni',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_trg_click_furni"}'::jsonb),
 (1000061,48973618,'wf_trg_user_clicks_tile','WIRED: User Clicks Floor Tile','floor',1,1,0.65,true,true,false,false,true,'wf_trg_user_clicks_tile',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_trg_click_furni"}'::jsonb),
 (1000062,90201279,'wf_trg_user_clicks_user','WIRED: User Clicks User','floor',1,1,0.65,true,true,false,false,true,'wf_trg_user_clicks_user',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_trg_click_user"}'::jsonb),
 (1000063,3681,'wf_act_reset_furni','WIRED: Reset Furniture State','floor',1,1,0.65,true,true,false,false,true,'wf_act_reset_furni',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_act_toggle_state"}'::jsonb),
 (1000064,32013636,'wf_cnd_user_performs_action','WIRED: User Performs Action','floor',1,1,0.65,true,true,false,false,true,'wf_cnd_user_performs_action',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_cnd_user_performs_action"}'::jsonb),
 (1000065,32013631,'wf_cnd_not_user_performs_action','WIRED: User Does Not Perform Action','floor',1,1,0.65,true,true,false,false,true,'wf_cnd_not_user_performs_action',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_cnd_not_user_performs_action"}'::jsonb),
 (1000066,98029998,'wf_cnd_not_has_handitem','WIRED: User Does Not Have Hand Item','floor',1,1,0.65,true,true,false,false,true,'wf_cnd_not_has_handitem',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_cnd_not_has_handitem"}'::jsonb),
 (1000067,32013633,'wf_cnd_team_has_rank','WIRED: Team Has Rank','floor',1,1,0.65,true,true,false,false,true,'wf_cnd_team_has_rank',4,'','',
  '{"source":"polaris-wired","wired_support":"native","asset_classname":"wf_cnd_team_has_rank"}'::jsonb),
 (1000068,3681,'wired_click_tile_anchor','WIRED Click Tile Anchor','floor',1,1,0.05,true,true,false,false,true,'room_invisible_click_tile',2,'','',
  '{"source":"pixels-wired-qa","asset_classname":"wf_act_toggle_state"}'::jsonb)
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

update furniture_definitions definition
set sprite_id=asset.sprite_id,interaction_type=asset.interaction,
 metadata=jsonb_set(
  definition.metadata,'{asset_classname}',to_jsonb(asset.classname),true),
 deleted_at=null,updated_at=now()
from (values
 ('wf_trg_user_clicks_furni',48973618,'wf_trg_click_furni'),
 ('wf_trg_user_clicks_tile',48973618,'wf_trg_click_furni'),
 ('wf_trg_user_clicks_user',90201279,'wf_trg_click_user'),
 ('wf_act_reset_furni',3681,'wf_act_toggle_state'),
 ('wf_cnd_user_performs_action',32013636,'wf_cnd_user_performs_action'),
 ('wf_cnd_not_user_performs_action',32013631,'wf_cnd_not_user_performs_action'),
 ('wf_cnd_not_has_handitem',98029998,'wf_cnd_not_has_handitem'),
 ('wf_cnd_team_has_rank',32013633,'wf_cnd_team_has_rank')
) as asset(interaction,sprite_id,classname)
where definition.name=asset.interaction;

insert into furniture_items(
 id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select fixture.id,definition.id,room.owner_player_id,room.id,
 fixture.x,fixture.y,fixture.z,fixture.rotation,fixture.extra_data
from (values
 (1010060,'wf_trg_user_clicks_furni',4,2,0,0,'0'),
 (1010061,'wf_trg_user_clicks_tile',5,2,0,0,'0'),
 (1010062,'wf_trg_user_clicks_user',6,2,0,0,'0'),
 (1010063,'wf_act_reset_furni',4,2,0,0,'0'),
 (1010064,'wf_cnd_user_performs_action',4,2,0,0,'0'),
 (1010065,'wf_cnd_not_user_performs_action',5,2,0,0,'0'),
 (1010066,'wf_cnd_not_has_handitem',6,2,0,0,'0'),
 (1010067,'wf_cnd_team_has_rank',7,2,0,0,'0'),
 (1020060,'wired_click_tile_anchor',6,6,0,0,'0'),
 (1020061,'wf_act_toggle_state',8,6,0,0,'1'),
 (1020064,'wf_act_reset_furni',5,2,0,0,'0'),
 (1020065,'wf_act_reset_furni',6,2,0,0,'0'),
 (1020066,'wf_trg_says_something',7,2,0,0,'0'),
 (1020067,'wf_act_reset_furni',7,2,0,0,'0')
) as fixture(id,definition_name,x,y,z,rotation,extra_data)
join furniture_definitions definition on definition.name=fixture.definition_name
 and definition.deleted_at is null
join rooms room on room.id=206 and room.name='WIRED Clicks and Actions'
on conflict(id) do nothing;

with desired(item_id,int_params,string_param,selection_mode,delay_pulses) as (values
 (1010060,'[100]','',1,0),
 (1010061,'[100]','',1,0),
 (1010062,'[]','',0,0),
 (1010063,'[]','',1,0),
 (1010064,'[10,0,0,1,1]','',0,0),
 (1010065,'[10,0,0,1,1]','',0,0),
 (1010066,'[1]','',0,0),
 (1010067,'[1,1]','',0,0),
 (1020064,'[]','',1,0),
 (1020065,'[]','',1,0),
 (1020066,'[]','rank',0,0),
 (1020067,'[]','',1,0)
)
insert into room_wired_settings(
 item_id,int_params,string_param,selection_mode,delay_pulses)
select desired.item_id,desired.int_params::jsonb,desired.string_param,
 desired.selection_mode,desired.delay_pulses
from desired
join furniture_items item on item.id=desired.item_id
where item.room_id=206 and item.deleted_at is null
on conflict(item_id) do update set int_params=excluded.int_params,
 string_param=excluded.string_param,selection_mode=excluded.selection_mode,
 delay_pulses=excluded.delay_pulses,version=room_wired_settings.version+1,
 updated_at=now();

with desired(
 wired_item_id,selected_item_id,ordinal,
 snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation) as (values
 (1010060,1020061,0,'1',8,6,0,0),
 (1010061,1020060,0,'0',6,6,0,0),
 (1010063,1020061,0,'1',8,6,0,0),
 (1020064,1020061,0,'1',8,6,0,0),
 (1020065,1020061,0,'1',8,6,0,0),
 (1020067,1020061,0,'1',8,6,0,0)
)
insert into room_wired_selected_items(
 wired_item_id,selected_item_id,ordinal,
 snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
select desired.*
from desired
join furniture_items wired on wired.id=desired.wired_item_id
join furniture_items selected on selected.id=desired.selected_item_id
where wired.room_id=206 and selected.room_id=206
 and wired.deleted_at is null and selected.deleted_at is null
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,
 snapshot_state=excluded.snapshot_state,snapshot_x=excluded.snapshot_x,
 snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,
 snapshot_rotation=excluded.snapshot_rotation;

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where room_id=206 and (id between 1010060 and 1020061 or id between 1020064 and 1020067); delete from furniture_definitions where id between 1000060 and 1000068 and metadata->>'source' in ('polaris-wired','pixels-wired-qa');
