--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0018-ensure-admin-wildcard context:development
insert into permission_group_nodes(group_id,node,allowed)
select id,'*',true
from permission_groups
where lower(name)='admin' and deleted_at is null
on conflict(group_id,node) do update set allowed=true;

--rollback select 1;
