--liquibase formatted sql

--changeset pixels:catalog-seed-wired-projectile-0036 context:development
insert into catalog_items(
 id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,
 limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value
select 1030068,1009,definition.id,'wf_xtra_projectile',3,0,-1,1,0,0,
 false,68,true,'0'
from furniture_definitions definition
where definition.name='wf_xtra_projectile' and definition.deleted_at is null
 and not exists (
  select 1 from catalog_items existing
  where existing.name='wf_xtra_projectile' and existing.deleted_at is null
 )
on conflict(id) do nothing;

select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id=1030068 and name='wf_xtra_projectile';
