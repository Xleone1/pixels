# Plugin Events: Commerce and World Objects

## Furniture placement

`furniture.place` runs after ownership, inventory state, marketplace state, and
trade-staging checks but before persistence. A listener can replace floor
coordinates, height, rotation, or wall position. Pixels validates the final
placement again; cancellation leaves the item in inventory.

All other furniture, pet, plant, and bot events are immutable realm facts
listed in [[PLUGINS-EVENTS-REALMS]].

## Catalog purchase

`catalog.purchase` runs after page/offer visibility and native free-price
calculation, but before transaction creation, currency deduction, or delivery.
Listeners can replace per-unit credit cost, point cost, and point type, or
cancel the purchase. Negative prices are rejected after callbacks.

Setting the final prices to zero is supported and still delivers the purchased
product. When no listener is active, the dispatcher is skipped so native free
mode and bulk-discount behavior remain unchanged.

Marketplace, crafting, recycler, groups, messenger, navigator, subscription,
guide, and moderation lifecycle events use immutable `event.Published`
snapshots. Their full stable-name catalog is in [[PLUGINS-EVENTS-REALMS]].
