--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0047-restore-gray-sofa context:development
update furniture_definitions
set interaction_type='default',
    effect_male=null,
    effect_female=null,
    allow_walk=false,
    updated_at=now()
where id=3 and name='sofa_silo';

--rollback update furniture_definitions set interaction_type='effect_tile',effect_male=90,effect_female=91,allow_walk=true,updated_at=now() where id=3 and name='sofa_silo';
