# Plugin Event: `furniture.move`

This mutable, cancellable event runs after native item and room checks and
before `MoveItem`. It exposes `Player`, `ItemID`, `RoomID`, and mutable `X`,
`Y`, `Z`, `Rotation`, and `WallPosition`.

```go
host.Events().Listen(event.FurnitureMoveName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	moved := current.(*event.FurnitureMove)
	if protectedRoom(moved.RoomID) {
		moved.SetCancelled(true)
	}
	return nil
})
```

Pixels validates the resulting placement again. Invalid callback coordinates
fail normally, and cancellation performs no persistence.
