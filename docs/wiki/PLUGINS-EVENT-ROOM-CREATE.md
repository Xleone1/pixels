# Plugin Event: `room.create`

`room.create` runs after initial native validation and layout lookup, but
before the room and tags are inserted. A listener may change `RoomName`,
`Description`, `ModelName`, `MaxUsers`, `CategoryID`, `TradeMode`, and `Tags`.

```go
host.Events().Listen(event.RoomCreateName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	created := current.(*event.RoomCreate)
	created.RoomName = strings.TrimSpace(created.RoomName)
	created.Tags = append(created.Tags, "plugin-managed")
	return nil
})
```

Pixels normalizes and validates all changed fields again, including resolving
the changed model and category. Cancellation inserts neither the room nor tags.
