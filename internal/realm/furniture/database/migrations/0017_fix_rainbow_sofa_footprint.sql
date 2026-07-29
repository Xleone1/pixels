--liquibase formatted sql

--changeset pixels:pixels-furniture-0017-fix-rainbow-sofa-footprint
update furniture_definitions
set width = 2,
    length = 1,
    updated_at = now(),
    version = version + 1
where name = 'nft_h23_rainbowsofa'
  and (width <> 2 or length <> 1);
