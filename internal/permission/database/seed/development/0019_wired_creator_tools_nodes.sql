--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0019-wired-creator-tools-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
select id,'room.wired.variables.manage',true
from permission_groups
where name='moderator'
on conflict(group_id,node) do update set allowed=excluded.allowed;
--rollback delete from permission_group_nodes where node='room.wired.variables.manage';
