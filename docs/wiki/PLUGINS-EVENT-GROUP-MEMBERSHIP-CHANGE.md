# Plugin Event: `group.membership.change`

This cancellable event covers group `join`, administrative `add`, and request
`accept` through `Action`. `Player`, `GroupID`, and `Role` are immutable
snapshots; `Role` is present for `add`.

```go
host.Events().Listen(event.GroupMembershipChangeName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	change := current.(*event.GroupMembershipChange)
	if exceedsWeeklyLimit(change.Player.ID) {
		change.SetCancelled(true)
	}
	return nil
})
```

Native authorization runs first where applicable. Cancellation prevents the
corresponding membership store method and projections.
