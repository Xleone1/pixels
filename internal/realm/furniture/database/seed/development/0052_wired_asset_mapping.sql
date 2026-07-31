--liquibase formatted sql

--changeset pixels:furniture-seed-wired-assets-0052 context:development
-- Map every expansion definition to a real, published FurnitureData sprite.
update furniture_definitions definition
set sprite_id=asset.sprite_id,
 metadata=jsonb_set(
  definition.metadata,'{asset_classname}',to_jsonb(asset.classname),true),
 updated_at=now()
from (values
 ('wf_slc_furni_area',966610621,'wf_slc_furni_area'),
 ('wf_slc_furni_neighborhood',857043807,'wf_slc_furni_neighborhood'),
 ('wf_slc_furni_bytype',446313468,'wf_slc_furni_bytype'),
 ('wf_slc_users_area',735519086,'wf_slc_users_area'),
 ('wf_slc_users_neighborhood',756632562,'wf_slc_users_neighborhood'),
 ('wf_slc_furni_altitude',252293444,'wf_slc_furni_altitude'),
 ('wf_slc_furni_onfurni',333117657,'wf_slc_furni_onfurni'),
 ('wf_slc_furni_picks',717724789,'wf_slc_furni_picks'),
 ('wf_slc_furni_signal',422001511,'wf_slc_furni_signal'),
 ('wf_slc_users_signal',451817134,'wf_slc_users_signal'),
 ('wf_slc_users_bytype',246616967,'wf_slc_users_bytype'),
 ('wf_slc_users_team',751055284,'wf_slc_users_team'),
 ('wf_slc_users_byaction',987510308,'wf_slc_users_byaction'),
 ('wf_slc_users_byname',807283715,'wf_slc_users_byname'),
 ('wf_slc_users_onfurni',352976345,'wf_slc_users_onfurni'),
 ('wf_slc_users_group',880755949,'wf_slc_users_group'),
 ('wf_slc_users_handitem',238800259,'wf_slc_users_handitem'),
 ('wf_slc_furni_with_var',674941925,'wf_slc_furni_with_var'),
 ('wf_slc_users_with_var',373654393,'wf_slc_users_with_var'),
 ('wf_slc_remote',2000028788,'wf_slc_remote'),
 ('wf_var_user',2000028784,'wf_var_user'),
 ('wf_var_furni',2000028787,'wf_var_furni'),
 ('wf_var_room',2000028790,'wf_var_room'),
 ('wf_var_reference',2000029123,'wf_var_reference'),
 ('wf_trg_recv_signal',2000029125,'wf_trg_recv_signal'),
 ('wf_trg_leave_room',123457967,'wf_trg_leave_room'),
 ('wf_trg_user_performs_action',4490566,'wf_trg_user_performs_action'),
 ('wf_trg_clock_counter',77201376,'wf_trg_clock_counter'),
 ('wf_trg_var_changed',32013641,'wf_trg_var_changed'),
 ('wf_act_send_signal',2000029124,'wf_act_send_signal'),
 ('wf_act_freeze',123457161,'wf_act_freeze'),
 ('wf_act_unfreeze',50380137,'wf_act_unfreeze'),
 ('wf_act_furni_to_user',123457970,'wf_act_furni_to_user'),
 ('wf_act_user_to_furni',5450,'wf_act_move_furni_to'),
 ('wf_act_furni_to_furni',2000025656,'wf_act_cnd_furni_to_furni'),
 ('wf_act_set_altitude',4490563,'wf_act_set_altitude'),
 ('wf_act_control_clock',2000025641,'wf_act_control_clock'),
 ('wf_act_control_clock_counter',23052801,'wf_act_adjust_clock'),
 ('wf_act_move_rotate_user',32013621,'wf_act_move_rotate_user'),
 ('wf_act_give_var',32013615,'wf_act_give_var'),
 ('wf_act_remove_var',32013622,'wf_act_remove_var'),
 ('wf_act_change_var_val',32013613,'wf_act_change_var_val'),
 ('wf_cnd_clock_matches',77201375,'wf_cnd_counter_time_matches'),
 ('wf_cnd_has_altitude',23437467,'wf_cnd_has_altitude'),
 ('wf_cnd_not_has_altitude',23437468,'wf_cnd_not_has_altitude'),
 ('wf_cnd_actor_dir',32013625,'wf_cnd_actor_dir'),
 ('wf_cnd_slc_quantity',32013632,'wf_cnd_slc_quantity'),
 ('wf_cnd_has_var',32013628,'wf_cnd_has_var'),
 ('wf_cnd_neg_has_var',32013629,'wf_cnd_neg_has_var'),
 ('wf_cnd_var_val_match',32013638,'wf_cnd_var_val_match'),
 ('wf_cnd_var_age_match',32013637,'wf_cnd_var_age_match'),
 ('wf_xtra_exec_in_order',77201377,'wf_xtra_exec_in_order'),
 ('wf_xtra_filter_furni_by_var',2000028765,'wf_xtra_filter_furni_by_var'),
 ('wf_xtra_filter_users_by_var',2000028766,'wf_xtra_filter_users_by_var')
) as asset(interaction,sprite_id,classname)
where definition.name=asset.interaction
 and definition.metadata->>'source'='polaris-wired';

--rollback update furniture_definitions set sprite_id=3681,metadata=metadata-'asset_classname',updated_at=now() where metadata->>'source'='polaris-wired';
