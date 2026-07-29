# Plugin Realm Event Catalog

SDK 3.x projects every committed realm fact by its stable internal event name.
Listen with `Host.Events().Listen(name, options, listener)` and type assert the
callback value to `*event.Published`. Field names match the exported payload
field names; `Fields()` and `Field()` always return detached values.

## Player, session, and chat

- Player: `player.authenticating`, `player.authenticated`,
  `player.authentication_failed`, `player.connected`,
  `player.disconnected`, `player.profile_loaded`,
  `player.effect.enabled`, `player.effect.expired`, `player.effect.granted`,
  `player.identity.name_changed`, `player.respect.granted`,
  `player.profile.updated`.
- Session: `session.bound`, `session.unbound`.
- Chat facts: `chat.mute_rejected`, `chat.shouted`, `chat.talked`,
  `chat.whispered`.
- Plugin-only detector: `command.attempted`.

`player.connected` and `inventory.currency_changed` use their richer typed SDK
values instead of `Published`.

## Rooms and world

- Access/runtime: `room.entered`, `room.left`, `room.occupancy_changed`.
- Control: `room.ambassador.alerted`, `room.moderation_banned`,
  `room.moderation_kicked`, `room.moderation_muted`,
  `room.moderation_unbanned`, `room.moderation_unmuted`,
  `room.floorplan_saved`, `room.mute_all_changed`, `room.rights_granted`,
  `room.rights_revoked`, `room.settings_updated`, `room.vote_cast`,
  `room.word_filter_modified`.
- Record: `room.bundle.purchased`, `room.created`, `room.deleted`,
  `room.staff_picked`, `room.updated`.
- World: `room.unit.danced`, `room.unit.expressed`,
  `room.unit.idle_changed`, `room.unit.moved`, `room.game.progressed`.

## Furniture, pets, and bots

- Furniture: `furniture.firework.charged`, `furniture.moved`,
  `furniture.picked_up`, `furniture.placed`, `furniture.postit_placed`,
  `furniture.random_resolved`, `furniture.rolled`,
  `furniture.surface_applied`, `furniture.teleport_completed`,
  `furniture.teleport_failed`, `furniture.teleport_started`,
  `furniture.used`, `furniture.walked_off`, `furniture.walked_on`.
- Pets/plants: `pet.breeding_completed`, `pet.harvested`, `plant.healed`,
  `plant.treated`, `pet.fed`, `pet.leveled`, `pet.respected`,
  `pet.dismounted`, `pet.mounted`, `pet.created`, `pet.deleted`,
  `pet.updated`, `pet.picked_up`, `pet.placed`.
- Bots: `bot.settings.chat_saved`, `bot.settings.look_saved`,
  `bot.settings.name_saved`, `bot.picked_up`, `bot.placed`,
  `bot.serve_item.requested`, `bot.shouted`, `bot.talked`,
  `bot.whispered`.

## Commerce and progression

- Catalog: `catalog.purchased`, `catalog.purchased.gift`,
  `catalog.voucher.redeemed`.
- Marketplace: `marketplace.expired`, `marketplace.listed`,
  `marketplace.sold`.
- Crafting: `exchange.redeemed`, `crafting.crafted`,
  `crafting.recipe.discovered`, `crafting.recipe.exhausted`,
  `recycler.recycled`.
- Subscription: `subscription.calendar.door_opened`,
  `subscription.created`, `subscription.expired`, `subscription.extended`,
  `subscription.club_gift.claimed`, `subscription.payday.awarded`.

## Social, moderation, and navigation

- Groups: `group.forum.posted`, `group.forum.thread.changed`,
  `group.created`, `group.deactivated`, `group.updated`,
  `group.membership.changed`, `group.favorite.changed`.
- Messenger: `messenger.relation.changed`, `messenger.friend.removed`,
  `messenger.request.accepted`, `messenger.request.declined`,
  `messenger.request.sent`, `messenger.player_ignored`,
  `messenger.invite.sent`, `messenger.message.sent`.
- Guide/moderation: `guide.enrolled`, `guide.session.completed`,
  `moderation.issue.created`, `moderation.issue.closed`,
  `sanction.applied`, `sanction.revoked`.
- Navigator: `navigator.search_executed`, `navigator.favorite_changed`,
  `navigator.closed`, `navigator.initialized`.
- Trades: `trade.started`, `trade.completed`, `trade.cancelled`.

The bridge test compares this runtime registry with all realm event constants,
so this catalog grows without silently dropping new server facts.
