# Plugin Event: `furniture.pickup`

This cancellable event runs after ownership and room-presence checks and before
the furniture returns to inventory. `Player`, `ItemID`, and `RoomID` are
immutable; there are no mutable destination fields.

```go
host.Events().Listen(event.FurniturePickupName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	pickup := current.(*event.FurniturePickup)
	pickup.SetCancelled(isPermanentExhibit(pickup.ItemID))
	return nil
})
```

Cancellation leaves the placed item unchanged.
