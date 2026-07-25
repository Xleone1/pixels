--liquibase formatted sql

--changeset pixels:pixels-pet-0009-extended-species
alter table pet_species drop constraint pet_species_type_id_check;
alter table pet_species add constraint pet_species_type_id_check check (type_id between 0 and 80);
alter table pet_product_rules drop constraint pet_product_rules_type_id_check;
alter table pet_product_rules add constraint pet_product_rules_type_id_check check (type_id between -1 and 80);

update pet_species
set slug='monster',display_key='pet.species.monster',behavior_kind='generic',
    breedable=true,enabled=true,version=version+1
where type_id=13;
update pet_breeds set sellable=true,enabled=true where type_id=13;

insert into pet_species(type_id,slug,display_key,behavior_kind,rideable,breedable,plant,enabled)
select type_id,'pet'||type_id,'pet.species.'||type_id,'generic',false,true,false,true
from generate_series(36,80) type_id
on conflict(type_id) do update set enabled=true,version=pet_species.version+1;

insert into pet_breeds(type_id,breed_id,palette_id,color,sellable,rarity,enabled)
select type_id,0,0,'FFFFFF',true,0,true from generate_series(36,80) type_id
on conflict(type_id,breed_id,palette_id) do update set sellable=true,enabled=true;

insert into pet_species_commands(type_id,command_id)
select species.type_id,command.id
from pet_species species cross join pet_commands command
where species.type_id between 36 and 80 and species.enabled and command.enabled
on conflict do nothing;

insert into pet_vocals(type_id,mood,text_key,weight,cooldown_ms)
select type_id,'idle','pet.vocal.generic',1,15000
from pet_species where type_id between 36 and 80 and enabled
on conflict do nothing;

insert into pet_breeding_rules(parent_one_type_id,parent_two_type_id,result_type_id)
select type_id,type_id,type_id
from pet_species where type_id between 36 and 80 and enabled and breedable
on conflict do nothing;

insert into pet_breeding_races(result_type_id,breed_id,palette_id,weight,mutation,enabled)
select type_id,breed_id,palette_id,greatest(1,100/(rarity+1)),rarity>0,enabled
from pet_breeds where type_id between 36 and 80 and enabled
on conflict do nothing;

--rollback delete from pet_breeding_races where result_type_id between 36 and 80; delete from pet_breeding_rules where result_type_id between 36 and 80; delete from pet_vocals where type_id between 36 and 80; delete from pet_species_commands where type_id between 36 and 80; delete from pet_breeds where type_id between 36 and 80; delete from pet_species where type_id between 36 and 80; update pet_species set slug='reserved13',display_key='pet.species.reserved13',behavior_kind='disabled',breedable=false,enabled=false,version=version+1 where type_id=13; update pet_breeds set sellable=false,enabled=false where type_id=13; alter table pet_product_rules drop constraint pet_product_rules_type_id_check; alter table pet_product_rules add constraint pet_product_rules_type_id_check check (type_id between -1 and 35); alter table pet_species drop constraint pet_species_type_id_check; alter table pet_species add constraint pet_species_type_id_check check (type_id between 0 and 35);
