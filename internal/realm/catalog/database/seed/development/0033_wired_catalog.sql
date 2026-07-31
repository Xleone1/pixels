--liquibase formatted sql

--changeset pixels:catalog-seed-wired-0033 context:development
insert into catalog_pages(
 id,parent_id,name,layout,icon_color,icon_image,min_rank,order_num,visible,enabled,club_only)
overriding system value values
 (1007,1000,'wired_selectors','default_3x3',1,9,1,7,true,true,false),
 (1008,1000,'wired_variables','default_3x3',1,9,1,8,true,true,false)
on conflict(id) do update set parent_id=excluded.parent_id,name=excluded.name,
 layout=excluded.layout,order_num=excluded.order_num,visible=true,enabled=true;

insert into catalog_items(
 id,page_id,definition_id,name,cost_credits,cost_points,points_type,amount,
 limited_stack,limited_sells,club_only,order_num,enabled,extra_data)
overriding system value
select 1030000 + row_number() over(order by definition.id) - 1,
 case
  when definition.name like 'wf_slc_%' then 1007
  when definition.name like 'wf_var_%' then 1008
  when definition.name like 'wf_trg_%' then 1001
  when definition.name like 'wf_act_%' then 1002
  when definition.name like 'wf_cnd_%' then 1003
  else 1004
 end,
 definition.id,definition.name,3,0,-1,1,0,0,false,
 row_number() over(order by definition.id),true,'0'
from furniture_definitions definition
where definition.metadata->>'source'='polaris-wired'
on conflict(id) do update set page_id=excluded.page_id,definition_id=excluded.definition_id,
 name=excluded.name,order_num=excluded.order_num,enabled=true;

select setval(pg_get_serial_sequence('catalog_pages','id'),greatest((select max(id) from catalog_pages),1));
select setval(pg_get_serial_sequence('catalog_items','id'),greatest((select max(id) from catalog_items),1));
--rollback delete from catalog_items where id between 1030000 and 1030053; delete from catalog_pages where id between 1007 and 1008;
