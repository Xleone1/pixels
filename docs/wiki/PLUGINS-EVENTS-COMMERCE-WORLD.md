# Plugin Events: Commerce and World Objects

## Furniture placement

`furniture.place` runs after ownership, inventory state, marketplace state, and
trade-staging checks but before persistence. A listener can replace floor
coordinates, height, rotation, or wall position. Pixels validates the final
placement again; cancellation leaves the item in inventory.

Movement and pickup now have equivalent pre-persistence gates; see
[[PLUGINS-EVENT-FURNITURE-MOVE]] and [[PLUGINS-EVENT-FURNITURE-PICKUP]].

## Catalog purchase

`catalog.purchase` runs after page/offer visibility and native free-price
calculation, but before transaction creation, currency deduction, or delivery.
Listeners can replace per-unit credit cost, point cost, and point type, or
cancel the purchase. Negative prices are rejected after callbacks.

Setting the final prices to zero is supported and still delivers the purchased
product. When no listener is active, the dispatcher is skipped so native free
mode and bulk-discount behavior remain unchanged.

Marketplace and crafting expose both immutable committed facts and focused
pre-commit gates. Their full stable-name catalog is in
[[PLUGINS-EVENTS-REALMS]].
