--liquibase formatted sql

--changeset pixels:pixels-furniture-0018-fix-multiheight-posture
update furniture_definitions
set interaction_type = 'multiheight',
    interaction_modes_count = 2,
    multiheight = '0.50;1.50',
    updated_at = now(),
    version = version + 1
where name = 'nft_c22_mvhqchair'
  and (
      interaction_type <> 'multiheight'
      or interaction_modes_count <> 2
      or multiheight <> '0.50;1.50'
  );

update furniture_definitions
set width = 2,
    length = 1,
    updated_at = now(),
    version = version + 1
where name = 'nft_u23_sonft_sofa'
  and (width <> 2 or length <> 1);

update furniture_definitions
set interaction_type = 'multiheight',
    updated_at = now(),
    version = version + 1
where name = 'Mut_Esquina'
  and interaction_type = 'multieheight';

update furniture_definitions
set multiheight = '1;0.75;0.50;0.25;0',
    updated_at = now(),
    version = version + 1
where name = 'black_corner'
  and multiheight = '1, 0.75, 0.50, 0.25, 0';
