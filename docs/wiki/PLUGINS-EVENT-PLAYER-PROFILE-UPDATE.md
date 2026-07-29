# Plugin Event: `player.profile.update`

This event intercepts motto and figure updates before player persistence.
`Motto`, `Figure`, and `Gender` are optional pointers: only fields participating
in the current operation are present.

```go
host.Events().Listen(event.PlayerProfileUpdateName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	update := current.(*event.PlayerProfileUpdate)
	if update.Motto != nil && blocked(*update.Motto) {
		update.SetCancelled(true)
	}
	return nil
})
```

Changed mottos are length-checked again. Changed figures are revalidated
against figure data, gender, unlocks, and active club entitlement.
