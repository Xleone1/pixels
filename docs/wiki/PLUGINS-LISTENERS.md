# Plugin Listeners and Interceptors

Pixels exposes two callback pipelines to plugins. Event listeners react to typed domain moments chosen by the host. Packet interceptors sit before native inbound packet handlers. They share priority, timeout, panic recovery, and plugin scope rules, but they are not interchangeable.

## Event listeners

A listener is registered by stable event name and receives an `sdk/event.Event`.
SDK 3.x exposes three shapes:

| Shape | Examples | Moment |
|---|---|---|
| Typed notification | `player.connected`, `inventory.currency_changed`, `command.attempted` | After a fact, or after a prefixed command is detected |
| Published realm fact | `room.created`, `trade.completed`, `messenger.message.sent` | After the owning realm publishes the committed fact |
| Cancellable/mutable event | `chat.send`, `currency.grant`, `room.update`, `catalog.purchase` | After native authorization and before mutation |

The event name selects the stream. The concrete type provides the data.

```go
err := host.Events().Listen(
	sdkevent.PlayerConnectedName,
	sdkevent.ListenerOptions{Priority: sdkplugin.PriorityNormal},
	func(ctx context.Context, current sdkevent.Event) error {
		connected, valid := current.(*sdkevent.PlayerConnected)
		if !valid {
			return nil
		}

		return host.Players().Message(connected.Player.ID, "Plugin ready")
	},
)
```

The internal realm bus is never exposed. A plugin can subscribe only to events the SDK deliberately projects and cannot publish arbitrary realm events.

## Immutable realm facts

Every event published below `internal/realm` is bridged by its stable name.
The callback receives a `*event.Published`; `Fields()` returns a detached map
and `Field(name)` returns one detached field. Composite values are normalized
to maps and slices, so no internal realm type crosses the SDK boundary.

```go
host.Events().Listen("room.created", event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	created, valid := current.(*event.Published)
	if !valid {
		return nil
	}
	roomID, _ := created.Field("RoomID")
	log.Printf("room created: %v", roomID)
	return nil
})
```

The bridge registry is checked by an architecture test against every
`const Name bus.Name` under `internal/realm`; adding an internal fact without
registering it for plugins fails the test. See [[PLUGINS-EVENTS-REALMS]] for
the complete name catalog.

## Command attempts

`command.attempted` detects every player message that starts with
`PIXELS_COMMAND_PREFIX`, before Brigadier resolves the root. It is emitted for
valid, malformed, incomplete, denied, and unknown commands. It is deliberately
not cancellable: command ownership and execution remain with the command tree.

```go
host.Events().Listen(event.CommandAttemptName, event.ListenerOptions{
	Priority: plugin.PriorityMonitor,
}, func(_ context.Context, current event.Event) error {
	attempt, valid := current.(*event.CommandAttempt)
	if valid {
		log.Printf("player=%d root=%q input=%q", attempt.Player.ID, attempt.Root, attempt.Input)
	}
	return nil
})
```

## Mutable and cancellable operations

Every pre-commit event owns `Clone()` and `Apply()`. Each callback receives its
own deep copy; Pixels commits that copy only when the callback returns
successfully. A timeout or panic cannot mutate the shared event later.

SDK 3.x mutable events are:

| Event | Mutable fields | Cancellation stops |
|---|---|---|
| `chat.send` | `Text` | Room delivery |
| `currency.grant` | `Amount` | Balance persistence |
| `room.update` | `Params` | Room settings persistence |
| `room.enter.attempt` | none | Runtime room admission |
| `room.unit.move` | `TargetX`, `TargetY` | Path acceptance |
| `room.moderation.action` | none | Local kick/mute/unmute/ban/unban |
| `sanction.apply` | `Reason`, `ExpiresAt` | Global sanction persistence/effect |
| `trade.start` | none | Session creation |
| `trade.confirm` | none | Item/currency settlement |
| `trade.cancel` | none | Session closure |
| `furniture.place` | coordinates, rotation, wall position | Furniture persistence |
| `furniture.move` | coordinates, rotation, wall position | Furniture movement |
| `furniture.pickup` | none | Inventory return |
| `catalog.purchase` | credit/point price and point type | Charge and delivery |
| `room.create` | name, description, model, capacity, category, trade mode, tags | Room and tag insert |
| `marketplace.list` | `RawPrice` | Token spend, reservation, and listing insert |
| `marketplace.buy` | `BuyerPrice` | Charge, transfer, and sold state |
| `player.profile.update` | optional motto, figure, and gender | Profile persistence |
| `bot.speech` | `Message` | Bot chat broadcast |
| `group.membership.change` | none | Join, add, or acceptance |
| `messenger.friend.request` | none | Friend request insert |
| `messenger.friend.accept` | none | Friendship insert |
| `crafting.craft` | `RewardDefinitionID` | Ingredient consumption and reward grant |

`chat.send` carries sanitized text. A listener may replace `Text`, cancel delivery, or do both.

```go
err := host.Events().Listen(
	sdkevent.ChatSendName,
	sdkevent.ListenerOptions{Priority: sdkplugin.PriorityHigh},
	func(_ context.Context, current sdkevent.Event) error {
		chat, valid := current.(*sdkevent.ChatSend)
		if !valid {
			return nil
		}

		if strings.EqualFold(chat.Text, "blocked phrase") {
			chat.SetCancelled(true)
			return nil
		}

		chat.Text = strings.ReplaceAll(chat.Text, "hello", "hi")
		return nil
	},
)
```

`IgnoreCancelled` skips a listener when an earlier callback has already cancelled the event. Leave it false when the listener must observe cancellation or may restore it.

## Priority order

Larger numeric values execute first. Registrations with the same priority keep registration order.

| Constant | Value | Intended use |
|---|---:|---|
| `PriorityHighest` | `200` | Earliest policy guard |
| `PriorityHigh` | `100` | Early mutation or validation |
| `PriorityNormal` | `0` | Ordinary behavior |
| `PriorityLow` | `-100` | Late behavior |
| `PriorityLowest` | `-200` | Final mutation |
| `PriorityMonitor` | `-1000` | Last, observation only |

A monitor should not mutate or uncancel an event. Its low value exists so it sees the final state produced by ordinary listeners.

## Packet interceptors

An interceptor receives an immutable player snapshot when authenticated, the inbound header, and a private copy of the encoded payload. A nil `Header` observes every inbound packet. A header pointer limits it to that one packet.

```go
err := host.Players().Intercept(
	func(ctx context.Context, packet sdkplugin.InterceptContext, next sdkplugin.Next) error {
		log.Printf("header=%d player=%d", packet.Header, packet.Player.ID)
		return next(ctx)
	},
	sdkplugin.InterceptOptions{Priority: sdkplugin.PriorityLow},
)
```

Calling `next(ctx)` advances to the next interceptor and eventually the native handler. Returning `nil` without calling `next` deliberately consumes the packet. Calling `next` more than once, after the callback has returned, or with a cancelled context is rejected.

An interceptor should not return an ordinary error to express a harmless cancellation. That error reaches the transport's protocol failure path. Consume intentionally with `return nil`, or continue with `return next(ctx)`.

## Isolation and deadlines

Every callback receives a context bounded by `PIXELS_PLUGIN_CALLBACK_TIMEOUT`. Code should stop work when `ctx.Done()` closes. Go cannot forcibly terminate a goroutine, so a plugin that ignores cancellation can still waste resources after the caller has moved on.

A recovered panic disables the complete plugin scope. Future listeners, commands, interceptors, and routes from that plugin stop running. A normal returned error is logged for listeners and does not disable the scope. A timed out listener has its late mutations discarded. A panicking or timed out interceptor that has not advanced the chain falls through to native handling so infrastructure failure does not silently swallow the player's packet.

Plugins execute inside the Pixels process. Recovery and deadlines isolate callbacks operationally, but they do not form a security sandbox. Only install binaries you trust.
