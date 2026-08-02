--liquibase formatted sql

--changeset pixels:room-seed-wired-clicks-actions-0021 context:development
with owner as (
 select id,username
 from players
 where deleted_at is null
 order by case when lower(username)='niflaot' then 0 else 1 end,id
 limit 1
)
insert into rooms(
 id,owner_player_id,owner_name,name,description,model_name,max_users,score,
 category_id,trade_mode,staff_picked)
overriding system value
select 206,owner.id,owner.username,'WIRED Clicks and Actions',
 'Floor tile, avatar click, action, hand item, team rank, and state reset fixtures.',
 'model_a',25,0,2,0,true
from owner
on conflict(id) do nothing;

insert into room_tags(room_id,tag)
select room.id,tag.name
from rooms room
cross join (values ('wired-clicks'),('wired-actions')) as tag(name)
where room.id=206 and room.name='WIRED Clicks and Actions'
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id=206 and tag in ('wired-clicks','wired-actions'); delete from rooms where id=206 and name='WIRED Clicks and Actions';
