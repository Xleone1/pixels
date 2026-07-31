--liquibase formatted sql

--changeset pixels:furniture-seed-wired-0050 context:development
-- Polaris-compatible WIRED boxes use the existing action definition protocol.
insert into furniture_definitions(
 id,sprite_id,name,public_name,kind,width,length,stack_height,allow_stack,allow_walk,
 allow_sit,allow_lay,allow_inventory_stack,interaction_type,interaction_modes_count,
 multiheight,custom_params,metadata)
overriding system value
select id + 90000,3681,name,'WIRED: ' || name,'floor',1,1,0.65,true,true,
 false,false,true,name,4,'','','{"source":"polaris-wired","wired_support":"native"}'
from (values
 (910000,'wf_slc_furni_area'),(910001,'wf_slc_furni_neighborhood'),
 (910002,'wf_slc_furni_bytype'),(910003,'wf_slc_users_area'),
 (910004,'wf_slc_users_neighborhood'),(910005,'wf_slc_furni_altitude'),
 (910006,'wf_slc_furni_onfurni'),(910007,'wf_slc_furni_picks'),
 (910008,'wf_slc_furni_signal'),(910009,'wf_slc_users_signal'),
 (910010,'wf_slc_users_bytype'),(910011,'wf_slc_users_team'),
 (910012,'wf_slc_users_byaction'),(910013,'wf_slc_users_byname'),
 (910014,'wf_slc_users_onfurni'),(910015,'wf_slc_users_group'),
 (910016,'wf_slc_users_handitem'),(910017,'wf_slc_furni_with_var'),
 (910018,'wf_slc_users_with_var'),(910019,'wf_slc_remote'),
 (910020,'wf_var_user'),(910021,'wf_var_furni'),
 (910022,'wf_var_room'),(910023,'wf_var_reference'),
 (910024,'wf_trg_recv_signal'),(910025,'wf_trg_leave_room'),
 (910026,'wf_trg_user_performs_action'),(910027,'wf_trg_clock_counter'),
 (910028,'wf_trg_var_changed'),(910029,'wf_act_send_signal'),
 (910030,'wf_act_freeze'),(910031,'wf_act_unfreeze'),
 (910032,'wf_act_furni_to_user'),(910033,'wf_act_user_to_furni'),
 (910034,'wf_act_furni_to_furni'),(910035,'wf_act_set_altitude'),
 (910036,'wf_act_control_clock'),(910037,'wf_act_control_clock_counter'),
 (910038,'wf_act_move_rotate_user'),(910039,'wf_act_give_var'),
 (910040,'wf_act_remove_var'),(910041,'wf_act_change_var_val'),
 (910042,'wf_cnd_clock_matches'),(910043,'wf_cnd_has_altitude'),
 (910044,'wf_cnd_not_has_altitude'),(910045,'wf_cnd_actor_dir'),
 (910046,'wf_cnd_slc_quantity'),(910047,'wf_cnd_has_var'),
 (910048,'wf_cnd_neg_has_var'),(910049,'wf_cnd_var_val_match'),
 (910050,'wf_cnd_var_age_match'),(910051,'wf_xtra_exec_in_order'),
 (910052,'wf_xtra_filter_furni_by_var'),(910053,'wf_xtra_filter_users_by_var')
) wired(id,name)
on conflict(id) do update set sprite_id=excluded.sprite_id,name=excluded.name,
 public_name=excluded.public_name,kind=excluded.kind,width=excluded.width,
 length=excluded.length,stack_height=excluded.stack_height,
 allow_stack=excluded.allow_stack,allow_walk=excluded.allow_walk,
 allow_sit=excluded.allow_sit,allow_lay=excluded.allow_lay,
 allow_inventory_stack=excluded.allow_inventory_stack,
 interaction_type=excluded.interaction_type,
 interaction_modes_count=excluded.interaction_modes_count,
 metadata=excluded.metadata,deleted_at=null,updated_at=now();

select setval(pg_get_serial_sequence('furniture_definitions','id'),greatest((select max(id) from furniture_definitions),1));
--rollback delete from furniture_definitions where metadata->>'source'='polaris-wired';
