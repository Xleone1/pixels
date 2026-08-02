--liquibase formatted sql

--changeset pixels:bot-seed-wired-projectile-obstacle-0004 context:development
with owner as (
 select id
 from players
 where deleted_at is null
 order by case when lower(username)='niflaot' then 0 else 1 end,id
 limit 1
)
insert into bots(
 owner_player_id,room_id,behavior_type,name,motto,figure,gender,
 x,y,z,rotation,can_walk,chat_auto,chat_random)
select owner.id,207,'generic','ProjectileUser','Stationary user collision target.',
 'hr-515-33.hd-600-1.ch-635-70.lg-695-82.sh-730-62','F',7,7,0,6,false,false,false
from owner
join rooms room on room.id=207 and room.name='WIRED QA Projectile'
where not exists (
 select 1 from bots where room_id=207 and name='ProjectileUser'
);

--rollback delete from bots where room_id=207 and name='ProjectileUser';
