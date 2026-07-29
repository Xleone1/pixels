# Plugin Event: `crafting.craft`

`crafting.craft` runs inside the crafting transaction after ingredients are
selected and before limited stock or inventory is consumed.
`RewardDefinitionID` is mutable.

```go
host.Events().Listen(event.CraftingCraftName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	craft := current.(*event.CraftingCraft)
	if craft.RecipeID == seasonalRecipe {
		craft.RewardDefinitionID = seasonalReward
	}
	return nil
})
```

The changed definition is used for the grant, definition lookup, result, and
post-commit fact. Cancellation consumes nothing.
