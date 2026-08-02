--liquibase formatted sql

--changeset pixels:room-seed-wired-selector-variable-0023 context:development
with owner as (
 select id,username
 from players
 where deleted_at is null
 order by case when lower(username)='niflaot' then 0 else 1 end,id
 limit 1
), desired(id,name,description,tag) as (values
 (200,'WIRED QA Selectors','Dedicated dynamic user and furniture selector fixtures.','wired-selectors'),
 (201,'WIRED QA Variables','Dedicated user, furniture, room, and reference variable fixtures.','wired-variables')
)
insert into rooms(
 id,owner_player_id,owner_name,name,description,model_name,max_users,score,
 category_id,trade_mode,staff_picked)
overriding system value
select desired.id,owner.id,owner.username,desired.name,desired.description,
 'model_5',25,0,2,0,true
from desired cross join owner
on conflict(id) do update set owner_player_id=excluded.owner_player_id,
 owner_name=excluded.owner_name,name=excluded.name,description=excluded.description,
 model_name=excluded.model_name,max_users=excluded.max_users,
 category_id=excluded.category_id,trade_mode=excluded.trade_mode,
 staff_picked=excluded.staff_picked,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag)
select desired.id,desired.tag
from (values (200,'wired-selectors'),(201,'wired-variables')) desired(id,tag)
join rooms room on room.id=desired.id
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback update rooms set name=case id when 200 then 'WIRED Selectors' else 'WIRED Variables' end,model_name='model_a' where id in (200,201);
