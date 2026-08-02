--liquibase formatted sql

--changeset pixels:catalog-seed-wired-clicks-actions-0035 context:development
with desired(id,definition_name,name,order_num) as (values
 (1030060,'wf_trg_user_clicks_furni','wf_trg_user_clicks_furni',60),
 (1030061,'wf_trg_user_clicks_tile','wf_trg_user_clicks_tile',61),
 (1030062,'wf_trg_user_clicks_user','wf_trg_user_clicks_user',62),
 (1030063,'wf_act_reset_furni','wf_act_reset_furni',63),
 (1030064,'wf_cnd_user_performs_action','wf_cnd_user_performs_action',64),
 (1030065,'wf_cnd_not_user_performs_action','wf_cnd_not_user_performs_action',65),
 (1030066,'wf_cnd_not_has_handitem','wf_cnd_not_has_handitem',66),
 (1030067,'wf_cnd_team_has_rank','wf_cnd_team_has_rank',67)
)
insert into catalog_items(
 id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,
 limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value
select desired.id,1009,definition.id,desired.name,3,0,-1,1,0,0,false,
 desired.order_num,true,'0'
from desired
join furniture_definitions definition on definition.name=desired.definition_name
 and definition.deleted_at is null
where not exists (
 select 1 from catalog_items existing
 where existing.name=desired.name and existing.deleted_at is null
)
on conflict(id) do nothing;

select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id between 1030060 and 1030067 and name in ('wf_trg_user_clicks_furni','wf_trg_user_clicks_tile','wf_trg_user_clicks_user','wf_act_reset_furni','wf_cnd_user_performs_action','wf_cnd_not_user_performs_action','wf_cnd_not_has_handitem','wf_cnd_team_has_rank');
