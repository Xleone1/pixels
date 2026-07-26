--liquibase formatted sql

--changeset pixels:furniture-seed-fix-wired-selection-modes-0048 context:development
update room_wired_settings
set selection_mode=1,updated_at=now()
where item_id in (426000,426010,426020,426030)
  and selection_mode=0
  and exists (
      select 1
      from room_wired_selected_items selected
      where selected.wired_item_id=room_wired_settings.item_id
  );

--rollback update room_wired_settings set selection_mode=0,updated_at=now() where item_id in (426000,426010,426020,426030);
