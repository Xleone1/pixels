# Plugin Event: `marketplace.buy`

`marketplace.buy` runs inside the listing transaction after the open listing is
locked and before the buyer is charged. `BuyerPrice` is mutable and already
includes native commission.

```go
host.Events().Listen(event.MarketplaceBuyName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	purchase := current.(*event.MarketplaceBuy)
	purchase.BuyerPrice = purchase.BuyerPrice * 80 / 100
	return nil
})
```

The final non-negative value is the amount actually charged and published in
the committed sale fact. Cancellation rolls back charge, transfer, and sold
state.
