--liquibase formatted sql

--changeset pixels:catalog-seed-0031-room-ads-recovery context:development
update catalog_pages
set parent_id=(select id from catalog_pages where name='vip' and deleted_at is null limit 1),
    layout='roomads',visible=true,enabled=true,deleted_at=null,updated_at=now(),version=version+1
where id=990;

update catalog_items
set enabled=true,deleted_at=null,updated_at=now(),version=version+1
where id=990001;

update catalog_pages
set layout='info_loyalty',updated_at=now(),version=version+1
where name='front_page' and deleted_at is null;
--rollback update catalog_pages set layout='room_ads',updated_at=now(),version=version+1 where id=990; update catalog_pages set layout='pets3',updated_at=now(),version=version+1 where name='front_page' and deleted_at is null;
