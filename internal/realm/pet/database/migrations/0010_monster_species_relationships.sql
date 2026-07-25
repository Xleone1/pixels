--liquibase formatted sql

--changeset pixels:pixels-pet-0010-monster-species-relationships
insert into pet_species_commands(type_id,command_id)
select 13,id from pet_commands where enabled
on conflict do nothing;

insert into pet_vocals(type_id,mood,text_key,weight,cooldown_ms)
select 13,'idle','pet.vocal.generic',1,15000
where not exists(select 1 from pet_vocals where type_id=13 and enabled);

insert into pet_breeding_rules(parent_one_type_id,parent_two_type_id,result_type_id,enabled)
values(13,13,13,true)
on conflict(parent_one_type_id,parent_two_type_id) do update set result_type_id=excluded.result_type_id,enabled=true;

insert into pet_breeding_races(result_type_id,breed_id,palette_id,weight,mutation,enabled)
select type_id,breed_id,palette_id,greatest(1,100/(rarity+1)),rarity>0,true
from pet_breeds where type_id=13 and enabled
on conflict(result_type_id,breed_id,palette_id) do update set weight=excluded.weight,mutation=excluded.mutation,enabled=true;

--rollback delete from pet_breeding_races where result_type_id=13; delete from pet_breeding_rules where parent_one_type_id=13 and parent_two_type_id=13; delete from pet_vocals where type_id=13; delete from pet_species_commands where type_id=13;
