# WIRED

Pixels implements the official WIRED behavior families using the
Polaris-compatible layout codes transported by Nitro's existing action,
trigger, and condition definition packets.

The immutable room generation contains 138 canonical behaviors:

- 25 triggers
- 44 effects
- 37 conditions
- 20 selectors
- 4 variable definitions
- 7 stack add-ons
- 1 highscore family

Selectors resolve only from warmed live-room state. Empty selector results are
authoritative and never fall back to an effect's stored targets. Remote
selectors are bounded to four levels, target results are deduplicated and
stable, and each target family is limited by
`PIXELS_WIRED_MAX_SELECTION_RESOLUTION`.

Durable variable assignments use four scopes: furniture, user, room, and
reference.
Rows are loaded once when the room runtime starts, read from the warmed cache,
written durably through `room_wired_variables`, and released from memory when
the room closes. Room deletion removes assignments through the database foreign
key.

## System and context variables

System variables begin with `@`, are derived from warmed room state, and are
read only. Furniture exposes identity, owner, sprite, position, dimensions,
posture capabilities, stackability, direction, height, and state. Users expose
identity, username, position, direction, hand item, effect, and idle state.
Rooms expose identity, owner, occupancy, capacity, and grid dimensions. WIRED
conditions resolve these values directly without a PostgreSQL read and without
building a temporary map.

Context is an ephemeral fifth scope and is not the durable reference scope. It
contains the latest execution event identity, actor, target, source furniture,
message, action, direction, signal, counters, scores, team, variable mutation,
and resolved actor or furniture counts. Context values exist only for the
execution that produced them. They are never inserted into
`room_wired_variables` and cannot be changed by an effect or Creator Tools.

## Creator Tools

The `:wired` and `:wf` commands open the compact room Creator Tools window on
Monitor. `:var` opens Variables and `:inspect` opens Inspection. The Monitor tab
shows compiled furniture, effects and stacks, outstanding delayed actions,
durable variables, the latest execution time, and the bounded trace ring. Its
maximum values combine room ceilings with per trace execution budgets, so a
compiled count and the adjacent execution limit do not always measure the same
thing. The latest completed trace retains its accepted signal count while the
configured signal budget remains visible. The Variables tab searches durable,
room system, and context values by scope. Inspection
selects a live furniture item or user and exposes that entity's complete system
variable set. Nitro uses the ordinary room selection arrow when hidden WIRED
boxes are disabled. Settings controls process local live refresh and the room's
durable hidden WIRED box preference. Live refresh returns to its default after
the process restarts.

`room.wired.inspect` grants read access and
`room.wired.variables.manage` grants variable mutations. Room furniture
managers and `room.wired.configure.any` retain their ordinary configuration
authority. The server always revalidates the active room, entity, scope,
permission, value bounds, and read only names.

Creator Tools is a Pixels protocol extension. Client packets `6300` through
`6303` request a snapshot, mutate a variable, inspect an entity, and update
settings. Server packets in the same directional range return versioned JSON
documents and stable result codes. Documents are bounded to 60 KiB, 512
entities, and 1024 variables. Variable integer values, scope identifiers, event
identifiers, and revisions use decimal strings so exact 64 bit values survive
JSON without JavaScript rounding. Creator Tools entity and variable scope ids
also travel as decimal strings. Room ids and the avatar unit id used by the
separate click packet remain signed 32 bit protocol fields and are validated
before use. Live refresh is requested by Nitro only while the panel is open.

Signals are intra-room, derived inside one serialized execution trace, and
discarded after the trace. `PIXELS_WIRED_MAX_SIGNALS_PER_TICK` prevents signal
cycles from consuming an unbounded event budget.

## Click triggers

Furniture clicks, click tile interactions, and avatar clicks are distinct
events. `wf_trg_user_clicks_furni` reacts to a deliberate furniture use.
`wf_trg_user_clicks_tile` reacts only when that furniture definition uses
`room_invisible_click_tile`; the same click still remains a normal furniture
click so existing stacks keep working. Both triggers support the trigger box,
explicit selected furniture, and selector source modes.

Nitro has no stock packet for a deliberate avatar click. Pixels reserves client
packet `6304` with one signed 32 bit room unit id. The server resolves that id
inside the player's active room, accepts another player unit only, and publishes
`wf_trg_user_clicks_user` after validation. A forged id, bot id, pet id, or
self click is ignored without exposing another room's state.

## Action and game conditions

`wf_cnd_user_performs_action` and `wf_cnd_not_user_performs_action` evaluate the
actor's authoritative room action. A triggering action pulse is checked first.
Persistent actions such as dancing, sitting, lying, and idle state can also be
read from the live unit. Optional sign and dance values restrict the condition
to one variant. Signs are pulses and never remain true after their event.

`wf_cnd_not_has_handitem` is the fail closed inverse of the existing hand item
condition. It does not pass when the player is missing. `wf_cnd_team_has_rank`
uses competition ranking among participating teams. Tied teams share a rank and
the following rank skips the occupied positions, so scores 10, 10, and 5 yield
ranks 1, 1, and 3.

## Furniture state reset

`wf_act_reset_furni` restores selected furniture state to `0` through the normal
durable compare and swap state service, then rebuilds and broadcasts the live
projection. It does not move or rotate furniture and does not depend on a saved
snapshot. This is intentionally separate from `wf_act_match_to_sshot`, which
restores the explicitly captured state and placement.

The reset effect uses effect client code `40`. Its current catalog definition
reuses the published toggle state visual because the reference asset set has no
dedicated reset box. The behavior key and server execution remain independent.

## Custom projectile

`wf_xtra_projectile`, client code `99`, is an explicit Pixels compatibility
extension rather than an official WIRED 2.0 piece. Its editor accepts direction
from `0` through `7`, travel distance from `1` through the configured
`PIXELS_WIRED_MAX_PROJECTILE_DISTANCE`, and room pulses per tile from `1`
through `20`. Saving and compiling reject a distance or total lifetime beyond
the environment limits instead of silently changing the requested behavior.
The selected furniture moves at most one tile in each eligible room cycle.
Swept footprint checks stop at room geometry, another furniture item, or
another room unit, so a fast configuration cannot skip a collision.

One bounded room state owns every active flight. The implementation creates no
goroutine or timer per projectile, broadcasts normal transient furniture and
height map packets, and persists only the final authoritative placement. A
persistence failure restores the original furniture and rider projections.
When the triggering actor is valid, Nitro follows that actor visually on the
projectile without creating another occupied room slot.

The immutable furniture variables
`@projectile.animation.tiles_travelled`,
`@projectile.animation.user_collisions`,
`@projectile.animation.furni_collisions`,
`@projectile.animation.position.x`,
`@projectile.animation.position.y`,
`@projectile.animation.altitude`,
`@projectile.animation.moving`, and `@can_move_freely` expose the current or
most recently completed flight. Environment limits bound active projectiles,
distance, duration, and the general compiled WIRED count.

Development rooms `200` and `201`, `WIRED QA Selectors` and `WIRED QA
Variables`, separate the selector and variable laboratories on large models.
Room `206`, `WIRED Clicks and Actions`, provides additive fixtures for click,
action, rank, hand item, and reset checks. Room `207`, `WIRED QA Projectile`,
provides two launch lanes, selected targets, furniture obstacles, and a
stationary bot obstacle for motion, collision, persistence, rider, and variable
checks. Every new QA seed prefers the player `Niflaot` as owner and uses a clean
database fallback only when that account is absent.

## Performance checks

Run the hot-path benchmarks with allocation reporting:

```sh
go test ./internal/realm/room/world/wired/runtime/tests \
  -run '^$' -bench 'Benchmark(NoCandidateDispatch|IndexedDispatch)$' -benchmem
go test ./internal/realm/room/world/wired/selection \
  -run '^$' -bench BenchmarkResolveArea -benchmem
go test ./internal/realm/room/world/wired/variable \
  -run '^$' -bench BenchmarkWarmedGet -benchmem
go test ./internal/realm/room/world/wired/projectile \
  -run '^$' -bench BenchmarkProjectileIdleCycle -benchmem
```

The no candidate event path, warmed variable lookup, and inactive projectile
cycle are required to remain allocation free.
