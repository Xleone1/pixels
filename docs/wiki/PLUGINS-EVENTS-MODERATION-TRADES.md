# Plugin Events: Moderation and Trades

## Global sanctions

`sanction.apply` runs after request normalization and authorization and before
the record or player effect. It covers warn, mute, kick, ban, and trade-lock.
Listeners may replace `Reason` and `ExpiresAt` or cancel the sanction. Pixels
normalizes and validates both fields again before persistence.

A cancelled warning creates no record and publishes no `sanction.applied`.
A cancelled ban cannot disconnect the target because side effects run only
after persistence.

## Local room moderation

`room.moderation.action` is distinct from global sanctions. It observes an
already-authorized local action and may veto kick, mute, unmute, ban, or unban
before the room moderation store changes.

## Trades

The three irreversible boundaries are independently cancellable:

- `trade.start` runs before the live session enters the registry.
- `trade.confirm` runs only when both participants have confirmed and before
  furniture/currency settlement. Cancellation resets confirmation ownership.
- `trade.cancel` runs before the registry closes an active session.

`Host.Trades().Active(playerID)` returns a detached snapshot.
`ForceCancel(playerID, reason)` closes an active trade through the same
cancellable boundary and records the plugin scope in the reason.

Committed lifecycle facts remain available as `trade.started`,
`trade.completed`, and `trade.cancelled`.
