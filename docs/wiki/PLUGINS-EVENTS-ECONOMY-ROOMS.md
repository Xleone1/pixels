# Plugin Events: Economy and Rooms

## Currency actions and interception

`Host.Economy()` exposes configured currency types and bounded balance
operations. `Grant` accepts a signed delta; `Set` accepts a non-negative
absolute balance. Both run through native validation. Plugin-originated
mutations use actor `plugin` and the scoped audit reason
`plugin:<plugin-name>:grant` or `plugin:<plugin-name>:set`.

Listen to `currency.grant` to cap or veto any signed grant before persistence:

```go
host.Events().Listen(event.CurrencyGrantName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	grant := current.(*event.CurrencyGrant)
	if grant.Amount > 100 {
		grant.Amount = 100
	}
	return nil
})
```

`inventory.currency_changed` is a typed post-commit notification with the
resulting balance, delta, currency type, actor kind, and player snapshot.

## Room actions and interception

`Host.Rooms()` reads immutable room snapshots, lists active occupant ids,
updates validated settings with optimistic locking, and controls active-room
mute-all state.

- `room.create` runs after native layout and content validation and before the
  room and its tags are inserted. See [[PLUGINS-EVENT-ROOM-CREATE]].
- `room.update` can replace any optional `Params` field or cancel persistence.
  Pixels reapplies native normalization, category policy, profanity policy,
  password hashing, and value validation after callbacks.
- `room.enter.attempt` runs after native access authorization but before the
  player or active room is mutated.
- `room.unit.move` can redirect `TargetX`/`TargetY` or cancel the movement.
  The redirected point is validated again. The hot path checks
  `HasListeners` before constructing the event.
- `room.moderation.action` can veto authorized local
  kick/mute/unmute/ban/unban operations.

Use [[PLUGINS-EVENTS-REALMS]] for post-commit room facts.
