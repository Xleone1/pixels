--liquibase formatted sql

--changeset pixels:furniture-seed-wired-labs-0051 context:development
insert into furniture_items(
 id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value
select 1010000 + definition.ordinal,definition.id,1,
 200 + definition.ordinal / 9,
 4 + definition.ordinal % 9 % 8,2 + definition.ordinal % 9 / 8,0,0,'0'
from (
 select id,row_number() over(order by id) - 1 ordinal
 from furniture_definitions
 where metadata->>'source'='polaris-wired'
) definition
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=1,
 room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=0,rotation=0,
 extra_data='0',deleted_at=null;

insert into furniture_items(
 id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data)
overriding system value values
 (1020900,28,1,200,4,8,0,0,'0'),(1020901,28,1,201,4,8,0,0,'0'),
 (1020902,28,1,202,4,8,0,0,'0'),(1020903,28,1,203,4,8,0,0,'0'),
 (1020904,28,1,204,4,8,0,0,'0'),(1020905,28,1,205,4,8,0,0,'0'),
 (1020000,1000026,1,205,4,5,0,0,'0'),
 (1020001,1000029,1,205,4,5,0.65,0,'0'),
 (1020010,1000024,1,205,6,5,0,0,'0'),
 (1020011,1000009,1,205,6,5,0.65,0,'0'),
 (1020012,3681,1,205,6,5,1.30,0,'0')
on conflict(id) do update set definition_id=excluded.definition_id,owner_player_id=1,
 room_id=excluded.room_id,x=excluded.x,y=excluded.y,z=excluded.z,
 rotation=excluded.rotation,extra_data=excluded.extra_data,deleted_at=null;

insert into room_wired_settings(
 item_id,int_params,string_param,selection_mode,delay_pulses)
select item.id,
 case
  when definition.interaction_type in (
   'wf_slc_furni_area','wf_slc_furni_neighborhood',
   'wf_slc_users_area','wf_slc_users_neighborhood') then '[0,0,12,12]'::jsonb
  when definition.interaction_type in (
   'wf_slc_furni_altitude','wf_cnd_has_altitude',
   'wf_cnd_not_has_altitude') then '[0,100]'::jsonb
  when definition.interaction_type in (
   'wf_slc_users_bytype','wf_slc_users_team','wf_slc_users_byaction',
   'wf_trg_user_performs_action','wf_trg_clock_counter',
   'wf_act_control_clock','wf_act_freeze') then '[1]'::jsonb
  when definition.interaction_type='wf_act_move_rotate_user' then '[0,0]'::jsonb
  when definition.interaction_type in (
   'wf_act_give_var','wf_act_change_var_val','wf_act_remove_var',
   'wf_cnd_has_var','wf_cnd_neg_has_var','wf_cnd_var_val_match',
   'wf_cnd_var_age_match') then '[3,1]'::jsonb
  when definition.interaction_type in (
   'wf_cnd_clock_matches','wf_cnd_slc_quantity') then '[0,1]'::jsonb
  when definition.interaction_type='wf_var_reference' then '[200]'::jsonb
  when definition.interaction_type in (
   'wf_xtra_filter_furni_by_var',
   'wf_xtra_filter_users_by_var') then '[0,0,1]'::jsonb
  else '[]'::jsonb
 end,
 case
  when definition.interaction_type in (
   'wf_trg_recv_signal','wf_act_send_signal') then 'wired'
  when definition.interaction_type='wf_slc_users_byname' then 'milo'
  when definition.interaction_type like 'wf_var_%'
    or definition.interaction_type like '%_var%'
    or definition.interaction_type in (
     'wf_trg_var_changed','wf_slc_furni_with_var',
     'wf_slc_users_with_var') then 'score'
  else ''
 end,
 case when definition.interaction_type in (
  'wf_slc_furni_bytype','wf_slc_furni_onfurni','wf_slc_furni_picks',
  'wf_slc_users_onfurni','wf_slc_remote','wf_act_furni_to_user',
  'wf_act_user_to_furni','wf_act_furni_to_furni','wf_act_set_altitude',
  'wf_cnd_has_altitude','wf_cnd_not_has_altitude') then 1 else 0 end,
 0
from furniture_items item
join furniture_definitions definition on definition.id=item.definition_id
where item.id between 1010000 and 1010053
on conflict(item_id) do update set int_params=excluded.int_params,
 string_param=excluded.string_param,selection_mode=excluded.selection_mode,
 delay_pulses=excluded.delay_pulses,version=room_wired_settings.version+1,
 updated_at=now();

insert into room_wired_settings(
 item_id,int_params,string_param,selection_mode,delay_pulses) values
 (1020000,'[1]','',0,0),(1020001,'[]','wired-integration',0,0),
 (1020010,'[]','wired-integration',0,0),(1020011,'[]','',0,0),
 (1020012,'[]','WIRED signal pipeline passed',0,0)
on conflict(item_id) do update set int_params=excluded.int_params,
 string_param=excluded.string_param,selection_mode=excluded.selection_mode,
 delay_pulses=excluded.delay_pulses,version=room_wired_settings.version+1,
 updated_at=now();

insert into room_wired_selected_items(
 wired_item_id,selected_item_id,ordinal,
 snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation)
select item.id,
 case when definition.interaction_type='wf_slc_remote' then 1010018
 else 1020900 + item.room_id - 200 end,
 0,'0',4,8,0,0
from furniture_items item
join furniture_definitions definition on definition.id=item.definition_id
where item.id between 1010000 and 1010053
 and definition.interaction_type in (
  'wf_slc_furni_bytype','wf_slc_furni_onfurni','wf_slc_furni_picks',
  'wf_slc_users_onfurni','wf_slc_remote','wf_act_furni_to_user',
  'wf_act_user_to_furni','wf_act_furni_to_furni','wf_act_set_altitude',
  'wf_cnd_has_altitude','wf_cnd_not_has_altitude')
on conflict(wired_item_id,selected_item_id) do update set ordinal=excluded.ordinal,
 snapshot_state=excluded.snapshot_state,snapshot_x=excluded.snapshot_x,
 snapshot_y=excluded.snapshot_y,snapshot_z=excluded.snapshot_z,
 snapshot_rotation=excluded.snapshot_rotation;

--rollback delete from furniture_items where id between 1010000 and 1020905;
