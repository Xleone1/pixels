# Permission Groups and Resolution

Pixels has two unrelated concepts called groups. Social groups are player communities. Permission groups are operator controlled authorization roles. This page covers permission groups and the exact algorithm used to decide one dotted capability.

## Permission groups are not Habbo groups

The `internal/realm/group` realm implements Habbo style social groups. They have a name, badge, owner, members, administrators, a home room, furniture rights, and a forum. Joining one is player facing social behavior.

`internal/permission` implements hotel wide authorization. Its groups resemble roles such as `member`, `moderator`, and `admin`. They have a numeric weight, optional inheritance, node grants, a client security level projection, and an optional room effect. Players never browse or join these roles through the Habbo group interface.

| Concept | Social group | Permission group |
|---|---|---|
| Owner | `internal/realm/group` | `internal/permission` |
| Purpose | Community identity and collaboration | Server authorization and staff capabilities |
| Typical names | Fan club, builders, event team | Member, moderator, admin |
| Membership | Managed through social group flows | Managed through protected permission administration |
| Badge and forum | Yes | No |
| Dotted capability nodes | No | Yes |
| Numeric weight | No | Yes |

A player may belong to many groups of either kind. Membership in a social group never grants a permission node unless a separate feature explicitly consults that social group.

## Dotted permission nodes

A node names one concrete capability:

```text
room.doorbell.answer.any
crafting.altar.manage.any
moderation.guide.duty
plugin.hello-plugin.hello.use
```

Segments use lowercase ASCII letters, digits, underscores, or hyphens. Dots express a namespace, not an automatic inheritance tree. A stored wildcard creates prefix coverage only when `*` is the final complete segment.

| Stored grant | Query | Matches | Specificity |
|---|---|---:|---:|
| `*` | `room.doorbell.answer.any` | Yes | 0 |
| `room.*` | `room.doorbell.answer.any` | Yes | 1 |
| `room.doorbell.*` | `room.doorbell.answer.any` | Yes | 2 |
| `room.doorbell.answer.any` | `room.doorbell.answer.any` | Yes | 4 |
| `room.doorbell.*` | `room.doorbell` | No | Not applicable |
| `room.*.answer` | Any query | No | Invalid syntax |

Specificity is the number of fixed segments. Exact nodes therefore beat broader wildcards inside the same resolution scope.

Realms register concrete nodes in code at startup. This produces a catalog with the declaring package, optional Nitro perk name, and plugin descriptions. Persistence may store wildcards, but checks always ask for a concrete registered capability.

### Core command nodes

| Node | Capability |
|---|---|
| `admin.alert` | Send a direct player alert with `:alert` |
| `admin.halert` | Broadcast a hotel alert with `:halert` |
| `admin.about` | Read private build and plugin metadata with `:about` |
| `admin.trace` | Capture the issuing player's packet traffic with `:trace` |
| `admin.effect` | Select or clear the issuing player's owned avatar effect with `:effect` |

`PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR=true` provides one narrow exception: a player without `admin.effect` may execute `:effect 0` to clear their own active effect. It never permits selecting a nonzero effect.

## Registered permission catalog

The following 113 concrete nodes are registered by production code. This is the
canonical operator-facing inventory: tests and fixtures are intentionally
excluded. Whenever code adds, removes, or renames a node, this catalog must be
updated in the same change.

### Administration, CMS, and moderation

| Node | Capability |
|---|---|
| `admin.about` | Read private build and plugin metadata with `:about` |
| `admin.alert` | Send a direct player alert with `:alert` |
| `admin.effect` | Select or clear an owned avatar effect with `:effect` |
| `admin.halert` | Broadcast a hotel alert with `:halert` |
| `admin.trace` | Capture the issuing player's packet traffic with `:trace` |
| `cms.maintenance.bypass` | Enter the CMS while maintenance mode is active |
| `cms.maintenance.early_access.manage` | Grant and revoke maintenance early access |
| `cms.maintenance.manage` | Configure CMS maintenance windows |
| `cms.news.manage` | Create, edit, and remove CMS news |
| `cms.permissions.groups.create` | Create permission groups from the CMS |
| `cms.permissions.groups.members.manage` | Add and remove permission-group members |
| `cms.permissions.groups.nodes.manage` | Grant, deny, and remove group nodes |
| `cms.permissions.groups.update` | Update permission-group metadata |
| `cms.permissions.groups.view` | View permission groups and their details |
| `cms.store.packages.manage` | Manage CMS store packages |
| `cms.store.transactions.authorize` | Authorize CMS store transactions |
| `cms.store.transactions.view` | View CMS store transactions |
| `cms.users.update` | Update player profiles from the CMS |
| `cms.users.view` | View player administration and durable account details |
| `moderation.chatlog.read` | Read moderation chat logs |
| `moderation.guardian.duty` | Perform Guardian duty flows |
| `moderation.guide.duty` | Perform Guide duty flows |
| `moderation.issue.manage` | Claim, reply to, and close moderation issues |
| `moderation.room.override` | Use staff moderation controls in any room |
| `moderation.sanction.apply` | Apply and revoke supported sanctions |
| `moderation.sanction.ban` | Apply ban sanctions |
| `moderation.sanction.immune` | Reject sanctions targeting this player |
| `moderation.tool.access` | Open and receive the moderation tool projection |

### Catalog, economy, subscriptions, and progression

| Node | Capability |
|---|---|
| `catalog.admin.manage` | Manage catalog pages and offers |
| `catalog.admin.voucher.manage` | Manage catalog voucher definitions |
| `currency.admin.manage` | Mutate player currency through administration |
| `currency.economy.infinite` | Bypass player-originated currency deductions |
| `marketplace.admin.manage` | Perform marketplace administration |
| `progression.definitions.manage.any` | Manage progression definitions |
| `progression.perk.trade` | Receive Nitro's `TRADE` perk and use trading |
| `progression.player.override.any` | Override player progression |
| `progression.quest.manage.any` | Manage quest definitions and state |
| `subscription.admin.calendar.manage` | Manage subscription calendar rewards |
| `subscription.admin.club_offer.manage` | Manage club subscription offers |
| `subscription.admin.membership.grant` | Grant club memberships administratively |
| `subscription.admin.targeted_offer.manage` | Manage targeted subscription offers |
| `subscription.calendar.staff.bypass` | Bypass normal calendar availability gates |
| `trade.bypass_restrictions` | Bypass ordinary trade restrictions |
| `trade.moderation.lock` | Lock or unlock trading through moderation |

### Chat, messenger, and player profile

| Node | Capability |
|---|---|
| `chat.bubble.any` | Use chat bubbles otherwise unavailable to the player |
| `chat.filter.immune` | Bypass the hotel chat word filter |
| `chat.flood.immune` | Bypass chat flood throttling |
| `chat.length.unlimited` | Bypass the normal chat message length limit |
| `chat.whisper.observe.any` | Observe whispers between other room occupants |
| `messenger.follow.any` | Follow players without the ordinary relationship gate |
| `messenger.friends.unlimited` | Bypass the Messenger friend limit |
| `navigator.favorite.unlimited` | Bypass the Navigator favorite-room limit |
| `player.admin.effect.grant` | Grant or revoke player effects through administration |
| `player.hotel.ambassador` | Project the hotel-ambassador client capability |
| `profile.respect.unlimited` | Bypass the daily respect quota |

### Camera, crafting, games, and bots

| Node | Capability |
|---|---|
| `bot.any_room_owner` | Manage bots as though owning their room |
| `bot.place_anywhere` | Place bots without the ordinary placement restriction |
| `bot.unlimited` | Bypass the bot ownership and room limits |
| `camera.capture.use` | Capture and purchase camera photos |
| `camera.gallery.moderate.any` | Moderate any camera gallery publication |
| `camera.settings.manage.any` | Manage camera settings for any target |
| `crafting.altar.manage.any` | Manage any crafting altar |
| `crafting.player.override.any` | Override player crafting state |
| `crafting.recycler.manage.any` | Manage recycler configuration and state |
| `games.center.manage.any` | Manage Game Center definitions |
| `games.polls.manage.any` | Manage hotel polls |

### Social groups

| Node | Capability |
|---|---|
| `group.badge.manage.any` | Change any social group's badge |
| `group.create` | Create a social group |
| `group.delete.any` | Deactivate or restore any social group |
| `group.forum.manage.any` | Manage any social-group forum |
| `group.forum.moderate.any` | Moderate posts and threads in any group forum |
| `group.home_room.rebind` | Rebind a group's home room |
| `group.manage.any` | Edit any social group's identity and settings |
| `group.members.manage.any` | Add, remove, or change members in any group |
| `group.read.deactivated` | Read deactivated social groups |
| `group.roles.manage.any` | Manage roles in any social group |

### Pets

| Node | Capability |
|---|---|
| `pet.inventory.limit.bypass` | Bypass the pet inventory limit |
| `pet.lifecycle.manage` | Administratively manage pet lifecycle state |
| `pet.manage.any` | Manage pets owned by another player |
| `pet.move.any` | Move any pet regardless of ordinary room authority |
| `pet.place.any` | Place pets without the ordinary ownership restriction |
| `pet.respect.limit.bypass` | Bypass the daily pet-respect limit |
| `pet.room.limit.bypass` | Bypass the per-room pet limit |

### Rooms and furniture

| Node | Capability |
|---|---|
| `room.admin.bundle_template.manage` | Manage hidden room-bundle templates |
| `room.ambassador.alert` | Send room ambassador alerts |
| `room.branding.manage` | Configure public image URLs on compatible room branding furniture |
| `room.delete.any` | Delete any room |
| `room.doorbell.answer.any` | Answer the doorbell in any room |
| `room.enter.any` | Bypass ordinary room entry policy |
| `room.enter.full` | Enter a room at its normal occupancy limit |
| `room.floorplan.any.edit` | Edit the floor plan of any room |
| `room.floorplan.own.edit` | Edit an owned room's floor plan |
| `room.furniture.any.manage` | Place, move, use, or pick up furniture in any room |
| `room.insight.read` | Read room inventory counters and bounded live profiling |
| `room.moderation.any.ban` | Ban a player from any room |
| `room.moderation.any.kick` | Kick a player from any room |
| `room.moderation.any.mute` | Mute a player in any room |
| `room.moderation.own.ban` | Ban a player from an owned room |
| `room.moderation.own.kick` | Kick a player from an owned room |
| `room.moderation.own.mute` | Mute a player in an owned room |
| `room.moderation.policy.any.manage` | Change moderation policy in any room |
| `room.moderation.policy.own.manage` | Change moderation policy in an owned room |
| `room.navigator.media.manage` | Configure CMS-owned Navigator room media |
| `room.promotion.manage.any` | Manage any room promotion |
| `room.rights.any.grant` | Grant rights in any room |
| `room.rights.any.revoke` | Revoke rights in any room |
| `room.rights.own.grant` | Grant rights in an owned room |
| `room.rights.own.revoke` | Revoke rights in an owned room |
| `room.settings.any.manage` | Change settings for any room |
| `room.settings.own.manage` | Change settings for an owned room |
| `room.staffpick.manage` | Add and remove staff-picked rooms |
| `room.unkickable` | Prevent room kicks and kick-like furniture effects |
| `room.wired.admin` | Use administrative Wired behavior |
| `room.wired.compatibility.use` | Use Wired compatibility behavior |
| `room.wired.configure` | Configure Wired with ordinary room authority |
| `room.wired.configure.any` | Configure Wired in any room |
| `room.wired.inspect` | Inspect Wired configuration |
| `room.wired.reward.manage` | Configure Wired rewards |

Plugin nodes are registered dynamically as
`plugin.<plugin-name>.<local-node>`, so their concrete inventory depends on the
plugins loaded by the current process. Catalog page `required_node` values do
not create permissions: they must reference an already registered concrete node,
including a loaded plugin node. Wildcards such as `*`, `room.*`, and
`plugin.example.*` are stored grants and are therefore not entries in this
concrete-node catalog.

## Grants and denies

Both permission groups and individual players may store a node with `allowed=true` or `allowed=false`. A false grant is an explicit deny. Removing a grant is different from denying it: removal lets the resolver continue to another source, while a deny is a decision.

The complete order is:

1. Resolve direct player grants.
2. If any direct grant matches, use the most specific direct match and stop.
3. Load the player's active permission groups by descending weight, then ascending group id for equal weights.
4. Resolve the first group whose inheritance chain contains any matching grant.
5. Ignore every lower weight group after that first group decision.
6. Deny when no source contains a matching grant.

This makes a direct player override absolute. A direct `room.* = false` wins even if an admin group grants `* = true`. It also means a high weight group with a matching deny wins over all lower weight groups.

## Resolution inside one group

Each permission group may have one parent. The resolver walks the selected group, then its parent, then the next parent until the chain ends. It detects cycles and rejects a broken chain.

Candidates inside that chain are compared in this order:

1. More fixed node segments win.
2. When specificity ties, the grant nearest the selected child group wins.
3. When specificity and inheritance depth both tie, deny wins.

Specificity comes before inheritance distance. For example, a parent exact grant beats a child wildcard because the exact node describes the requested capability more precisely. A child exact deny beats a parent exact allow because both have the same specificity and the child is nearer.

## Worked examples

Assume a player belongs to `moderator` at weight 50 and `member` at weight 0. `moderator` inherits from `member`.

| Grants | Query | Result | Reason |
|---|---|---|---|
| Moderator has `room.* = true` | `room.doorbell.answer.any` | Allow | First matching group and prefix wildcard |
| Moderator has `room.* = false`, member has exact allow | `room.doorbell.answer.any` | Deny | Moderator already produced a decision, so member membership is not considered |
| Moderator child has `room.* = false`, inherited member has exact allow | `room.doorbell.answer.any` | Allow | Both are in one inheritance chain and the parent exact grant is more specific |
| Moderator child has exact deny, inherited member has exact allow | `room.doorbell.answer.any` | Deny | Same specificity, nearer child wins |
| Direct player exact allow, admin group has `* = false` | Same exact node | Allow | Direct player decisions are resolved before groups |
| No matching direct or group grant | Any concrete node | Deny | Permissions are closed by default |

## Weight and primary group

Weight determines which membership is considered first and which group is exposed as the player's primary permission group. It is not added together, and lower groups do not contribute after a higher group has made a matching decision.

The primary group is simply the active membership with greatest weight. Pixels uses it for client security projection and synthetic group effects. Authorization still follows the complete node algorithm, so being primary does not imply every capability.

The development seeds make `demo` an admin at weight 100, `alice` a moderator at weight 50, and `bob` plus `carol` members at weight 0. Every created player receives the default `member` membership atomically.

## Perks and live projection

A registered node may map to a Nitro perk name. `EffectivePerks` resolves every mapped node and sends only allowed perks through `USER_PERKS`. The client view is therefore derived from the server permission engine rather than maintained as a second authorization list.

Permission records use a local cache for hot checks and Redis for shared cache fragments. Mutations invalidate the affected player, membership, group, and node fragments. Online players receive refreshed permissions and perks after the database commit. The warmed local resolution path is designed to remain allocation free.

## Administration

The protected permission routes expose the registered catalog, permission groups, memberships, direct player grants, effective nodes, and individual checks. `GET /docs` in development documents the exact request bodies and responses.

The API key authenticates the HTTP caller to the private API boundary. Permission nodes then authorize the acting player for domain operations. These are separate checks: possession of `X-API-Key` should not be treated as an unlimited staff rank inside gameplay workflows.
