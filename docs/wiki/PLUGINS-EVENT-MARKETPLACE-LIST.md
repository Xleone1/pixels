# Plugin Event: `marketplace.list`

This mutable, cancellable event runs after the requested price passes native
limits and before a token is spent or furniture is reserved. `RawPrice` is the
seller price before commission.

```go
host.Events().Listen(event.MarketplaceListName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	listing := current.(*event.MarketplaceList)
	listing.RawPrice = max(listing.RawPrice, seasonalMinimum)
	return nil
})
```

The changed price is checked against configured minimum and maximum values.
Cancellation leaves the token and furniture untouched.
