# Plugin Event: `messenger.friend.request`

This cancellable event runs after player lookup, privacy, duplicate, and both
friend-capacity checks, immediately before request persistence.

```go
host.Events().Listen(event.FriendRequestName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	request := current.(*event.FriendRequest)
	request.SetCancelled(accountTooNew(request.Sender.ID))
	return nil
})
```

`Sender` and `Target` are immutable player snapshots. Cancellation does not
create the request.
