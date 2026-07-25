--liquibase formatted sql

--changeset pixels:catalog-seed-0030-catalog-compatibility context:development
update catalog_pages
set layout='roomads',updated_at=now(),version=version+1
where name='room_ads' and deleted_at is null and layout='room_ads';

update catalog_pages
set layout='info_loyalty',updated_at=now(),version=version+1
where name='front_page' and deleted_at is null and layout='pets3';
--rollback update catalog_pages set layout='room_ads',updated_at=now(),version=version+1 where name='room_ads' and deleted_at is null and layout='roomads'; update catalog_pages set layout='pets3',updated_at=now(),version=version+1 where name='front_page' and deleted_at is null and layout='info_loyalty';
