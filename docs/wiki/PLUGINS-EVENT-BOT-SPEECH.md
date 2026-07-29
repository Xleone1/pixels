# Plugin Event: `bot.speech`

`bot.speech` runs after native global and room word filters and before packet
encoding. `Message` is mutable; `Scope` is the immutable value `talk`, `shout`,
or `whisper`.

```go
host.Events().Listen(event.BotSpeechName, event.ListenerOptions{}, func(_ context.Context, current event.Event) error {
	speech := current.(*event.BotSpeech)
	speech.Message = strings.ReplaceAll(speech.Message, "{hotel}", "Pixels")
	return nil
})
```

Cancellation or an empty resulting message suppresses broadcast and the
post-delivery bot fact.
