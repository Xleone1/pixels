# Plugin Event: `messenger.friend.accept`

This cancellable event runs after both players and friend-capacity limits are
validated and before the request becomes a friendship.

```go
host.Events().Listen(event.FriendAcceptName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	accepted := current.(*event.FriendAccept)
	accepted.SetCancelled(blockedPair(accepted.Actor.ID, accepted.Requester.ID))
	return nil
})
```

Cancellation performs neither the friendship insert nor card-cache
invalidation.
