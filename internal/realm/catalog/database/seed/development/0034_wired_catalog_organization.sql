--liquibase formatted sql

--changeset pixels:catalog-seed-wired-0034 context:development
-- Keep the established category ids and introduce one level of functional grouping.
insert into catalog_pages(
 id,parent_id,name,layout,icon_color,icon_image,min_rank,order_num,
 visible,enabled,club_only)
overriding system value values
 (1010,1000,'wired_classic','default_3x3',1,9,1,1,true,true,false),
 (1100,1000,'wired_advanced','default_3x3',2,10,1,2,true,true,false),
 (1009,1100,'wired_advanced_components','default_3x3',2,11,1,3,true,true,false)
on conflict(id) do update set parent_id=excluded.parent_id,name=excluded.name,
 layout=excluded.layout,icon_color=excluded.icon_color,
 icon_image=excluded.icon_image,order_num=excluded.order_num,
 visible=true,enabled=true;

update catalog_pages
set parent_id=1010,
 order_num=case id when 1001 then 1 when 1002 then 2 when 1003 then 3
  when 1004 then 4 else 5 end,
 visible=true,enabled=true
where id between 1001 and 1005;

update catalog_pages
set parent_id=1100,
 order_num=case id when 1007 then 1 else 2 end,
 visible=true,enabled=true
where id in (1007,1008);

-- The expansion was already seeded once; reorganize those rows without duplicating them.
update catalog_items item
set page_id=case
 when definition.name like 'wf_slc_%' then 1007
 when definition.name like 'wf_var_%' then 1008
 else 1009
 end,
 updated_at=now()
from furniture_definitions definition
where item.definition_id=definition.id
 and definition.metadata->>'source'='polaris-wired';

select setval(
 pg_get_serial_sequence('catalog_pages','id'),
 greatest((select max(id) from catalog_pages),1));

--rollback update catalog_items item set page_id=case when definition.name like 'wf_slc_%' then 1007 when definition.name like 'wf_var_%' then 1008 when definition.name like 'wf_trg_%' then 1001 when definition.name like 'wf_act_%' then 1002 when definition.name like 'wf_cnd_%' then 1003 else 1004 end from furniture_definitions definition where item.definition_id=definition.id and definition.metadata->>'source'='polaris-wired';
--rollback update catalog_pages set parent_id=1000 where id between 1001 and 1005 or id in (1007,1008);
--rollback delete from catalog_pages where id in (1009,1010,1100);
