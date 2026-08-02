--liquibase formatted sql

--changeset pixels:room-seed-wired-projectile-0022 context:development
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
select 207,owner.id,owner.username,'WIRED QA Projectile',
 'Bounded projectile motion, collision, system variable, and rider projection fixtures.',
 'model_a',25,0,2,0,true
from owner
on conflict(id) do nothing;

insert into room_tags(room_id,tag)
select room.id,tag.name
from rooms room
cross join (values ('wired-projectile'),('wired-integration')) as tag(name)
where room.id=207 and room.name='WIRED QA Projectile'
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id=207 and tag in ('wired-projectile','wired-integration'); delete from rooms where id=207 and name='WIRED QA Projectile';
