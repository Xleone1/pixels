--liquibase formatted sql

--changeset pixels:catalog-seed-0032-complete-layout-support context:development
update catalog_pages set layout='vip_buy',updated_at=now(),version=version+1 where layout in ('club_buy','loyalty_vip_buy') and deleted_at is null;
update catalog_pages set layout='club_gifts',updated_at=now(),version=version+1 where layout='club_gift' and deleted_at is null;
update catalog_pages set layout='default_3x3',updated_at=now(),version=version+1 where layout in ('collectibles','petcustomization','recycler_prizes') and deleted_at is null;
update catalog_pages set layout='info_loyalty',updated_at=now(),version=version+1 where layout in ('frontpage','frontpage_featured','info_duckets','info_pets','info_rentables','recycler','recycler_info') and deleted_at is null;
update catalog_pages set layout='guild_custom_furni',updated_at=now(),version=version+1 where layout='guild_furni' and deleted_at is null;
update catalog_pages set layout='guild_frontpage',updated_at=now(),version=version+1 where layout='guilds' and deleted_at is null;
update catalog_pages set layout='default_3x3_color_grouping',updated_at=now(),version=version+1 where layout='plasto' and deleted_at is null;
update catalog_pages set layout='single_bundle',updated_at=now(),version=version+1 where layout='productpage1' and deleted_at is null;
update catalog_pages set layout='spaces_new',updated_at=now(),version=version+1 where layout='spaces' and deleted_at is null;
update catalog_pages set visible=false,enabled=false,updated_at=now(),version=version+1 where layout='soundmachine' and deleted_at is null;

update catalog_pages shop
set parent_id=badges.id,visible=true,enabled=true,deleted_at=null,updated_at=now(),version=shop.version+1
from catalog_pages badges
where shop.name='badge_shop' and badges.name='badges' and badges.deleted_at is null;

update catalog_pages set layout='bots',visible=true,enabled=true,deleted_at=null,updated_at=now(),version=version+1 where name='bots';

insert into catalog_items (id,page_id,definition_id,reward_kind,name,cost_credits,cost_points,points_type,amount,limited_stack,club_only,order_num,enabled,giftable,extra_data)
overriding system value
select source.id,page.id,null,'bot',source.name,source.cost,0,-1,1,0,false,source.order_num,true,false,source.extra_data
from catalog_pages page
cross join (values
    (990101,'bot_generic',25,1,'name:Robbie;motto:Generic;figure:hr-3020-31.sh-3089-82.lg-3057-1330.ch-225-FFFFFF.ca-3084-82-82.wa-2003-63.hd-3091-1383;gender:m'),
    (990102,'bot_bartender',40,2,'name:Love;motto:Bartender;figure:hr-9534-45.sh-3064-1425.lg-3058-82.ch-818-FFFFFF.wa-2005-63.hd-600-1;gender:f'),
    (990103,'rentable_bot_visitor_log',35,3,'name:Belle;motto:Visitor Counter;figure:sh-3064-91.lg-3166-82.ch-3076-82-73.hr-999999357-31-896018.hd-3096-1;gender:f')
) as source(id,name,cost,order_num,extra_data)
where page.name='bots' and page.deleted_at is null
  and not exists (select 1 from catalog_items item where item.page_id=page.id and item.name=source.name and item.deleted_at is null)
on conflict (id) do update set page_id=excluded.page_id,reward_kind='bot',definition_id=null,name=excluded.name,cost_credits=excluded.cost_credits,cost_points=0,points_type=-1,amount=1,limited_stack=0,club_only=false,order_num=excluded.order_num,enabled=true,giftable=false,extra_data=excluded.extra_data,deleted_at=null,updated_at=now();
--rollback delete from catalog_items where id between 990101 and 990103;
