# Plugin WIRED Behaviors

SDK 3 lets a plugin register server logic for a namespaced WIRED effect or
condition. It does not modify the frozen native registry and cannot add a new
Nitro editor screen. A plugin reuses one editor layout already shipped by
Nitro, declared through `wired.Descriptor`.

## Register an effect

Registration belongs in `Plugin.Register`. The key must start with
`plugin.<manifest-name>.`; wrong namespaces and duplicates fail startup
registration.

```go
err := host.Wired().RegisterEffect(
	"plugin.example-plugin.room_notice",
	func(ctx context.Context, node wired.Node, fired wired.Event) (wired.EffectResult, error) {
		if fired.PlayerID != 0 {
			if err := host.Players().Message(fired.PlayerID, node.Text); err != nil {
				return wired.EffectResult{}, err
			}
		}
		return wired.EffectResult{Status: wired.EffectApplied}, nil
	},
	wired.WithDescriptor(wired.Descriptor{
		ClientCode: 7,
		Selection:  wired.SelectionNone,
		Actor:      wired.ActorPlayer,
		Editor:     true,
	}),
)
```

Configure the backing furniture definition with the exact interaction
`plugin.example-plugin.room_notice`. The stock editor writes its text,
numeric values, target ids, and delay into the normal WIRED record. Pixels
resolves the plugin descriptor when opening, saving, compiling, and executing
that record.

## Register a condition

```go
err := host.Wired().RegisterCondition(
	"plugin.example-plugin.actor_allowed",
	func(ctx context.Context, node wired.Node, fired wired.Event) (bool, bool, error) {
		if fired.PlayerID == 0 {
			return false, false, nil
		}
		return allowed(fired.PlayerID, node.Text), true, nil
	},
	wired.WithDescriptor(wired.Descriptor{
		ClientCode: 11,
		Selection:  wired.SelectionNone,
		Actor:      wired.ActorPlayer,
		Editor:     true,
	}),
)
```

The two booleans mean `pass` and `valid`. An invalid condition fails closed.
One key represents one family, so an effect and condition cannot claim the
same key.

## Reusable stock editor layouts

`ClientCode` is family-specific. The most useful existing effect layouts are:

| Effect code | Existing Nitro editor shape | Typical descriptor |
|---:|---|---|
| `0` | Select furniture and toggle state | `SelectionRequired` |
| `1` | Reset timers, no parameters | `SelectionNone` |
| `4` | Move/rotate selected furniture | `SelectionRequired` |
| `7` | Free text / show message | `ActorPlayer` |
| `8` | Teleport to selected furniture | `SelectionRequired`, `ActorUnit` |
| `17` | Reward configuration | `ActorPlayer` |
| `18` | Select stacks to call | `SelectionRequired` |
| `23` | Bot name and speech text | `ActorOptional` |

Useful condition layouts are:

| Condition code | Existing Nitro editor shape | Typical descriptor |
|---:|---|---|
| `0` | Match selected furniture snapshot | `SelectionRequired` |
| `3` / `4` | Time threshold | `SelectionNone` |
| `5` | Minimum/maximum user count | `SelectionNone` |
| `11` | Badge code text | `ActorPlayer` |
| `12` | Effect id | `ActorPlayer` |
| `24` | Date range | `SelectionNone` |
| `25` | Hand-item id | `ActorPlayer` |

Use the code from the same family and verify it with the Nitro version deployed
by the hotel. `Editor: false` is appropriate only for records configured
outside Nitro.

## Callback data and results

`wired.Node` contains detached copies of `Targets` and `Values`, plus `Text`,
`Duration`, room id, item id, and key. `wired.Event` carries room, player, and
source-item ids. Plugins never receive mutable room, furniture, or engine
objects.

Effects return `EffectApplied`, `EffectBlocked`, or `EffectSkipped`.
`ResetTimers` requests the native timer reset behavior; `CallTargets` enqueues
the listed stack ids through the ordinary bounded execution engine.

Callbacks use `PIXELS_PLUGIN_CALLBACK_TIMEOUT` and the standard panic recovery.
A panic disables that plugin scope without stopping native WIRED or unrelated
plugins. Native `wf_*` keys always resolve before plugin fallbacks and retain
their frozen descriptors and execution paths.
