# WIRED

Pixels implements the official WIRED behavior families using the
Polaris-compatible layout codes transported by Nitro's existing action,
trigger, and condition definition packets.

The immutable room generation contains 130 canonical behaviors:

- 22 triggers
- 43 effects
- 33 conditions
- 20 selectors
- 4 variable definitions
- 7 stack add-ons
- 1 highscore family

Selectors resolve only from warmed live-room state. Empty selector results are
authoritative and never fall back to an effect's stored targets. Remote
selectors are bounded to four levels, target results are deduplicated and
stable, and each target family is limited by
`PIXELS_WIRED_MAX_SELECTION_RESOLUTION`.

Variable assignments use four scopes: furniture, user, room, and reference.
Rows are loaded once when the room runtime starts, read from the warmed cache,
written durably through `room_wired_variables`, and released from memory when
the room closes. Room deletion removes assignments through the database foreign
key.

Signals are intra-room, derived inside one serialized execution trace, and
discarded after the trace. `PIXELS_WIRED_MAX_SIGNALS_PER_TICK` prevents signal
cycles from consuming an unbounded event budget.

## Performance checks

Run the hot-path benchmarks with allocation reporting:

```sh
go test ./internal/realm/room/world/wired/runtime/tests \
  -run '^$' -bench 'Benchmark(NoCandidateDispatch|IndexedDispatch)$' -benchmem
go test ./internal/realm/room/world/wired/selection \
  -run '^$' -bench BenchmarkResolveArea -benchmem
go test ./internal/realm/room/world/wired/variable \
  -run '^$' -bench BenchmarkWarmedGet -benchmem
```

The no-candidate event path and warmed variable lookup are required to remain
allocation free.
