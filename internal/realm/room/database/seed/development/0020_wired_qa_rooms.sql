--liquibase formatted sql

--changeset pixels:room-seed-wired-0020 context:development
insert into rooms(
 id,owner_player_id,owner_name,name,description,model_name,max_users,score,
 category_id,trade_mode,staff_picked)
overriding system value values
 (200,1,'milo','WIRED Selectors','Dynamic user and furniture selector fixtures.','model_a',25,0,2,0,false),
 (201,1,'milo','WIRED Variables','Durable room, user, furniture and reference variables.','model_a',25,0,2,0,false),
 (202,1,'milo','WIRED Signals','Bounded signal pipelines and signal selectors.','model_a',25,0,2,0,false),
 (203,1,'milo','WIRED Actions','Movement, freeze, altitude and clock effects.','model_a',25,0,2,0,false),
 (204,1,'milo','WIRED Conditions','Clock, direction, quantity and variable conditions.','model_a',25,0,2,0,false),
 (205,1,'milo','WIRED Integration','End-to-end selector, condition and effect pipelines.','model_a',25,0,2,0,false)
on conflict(id) do update set owner_player_id=excluded.owner_player_id,
 owner_name=excluded.owner_name,name=excluded.name,description=excluded.description,
 model_name=excluded.model_name,max_users=excluded.max_users,
 category_id=excluded.category_id,trade_mode=excluded.trade_mode,
 staff_picked=false,deleted_at=null,updated_at=now();

insert into room_tags(room_id,tag) values
 (200,'wired-selectors'),(201,'wired-variables'),(202,'wired-signals'),
 (203,'wired-actions'),(204,'wired-conditions'),(205,'wired-integration')
on conflict do nothing;

select setval(pg_get_serial_sequence('rooms','id'),greatest((select max(id) from rooms),1));
--rollback delete from room_tags where room_id between 200 and 205; delete from rooms where id between 200 and 205;
