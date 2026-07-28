package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

// CommandAttemptName identifies a prefixed command attempt.
const CommandAttemptName = "command.attempted"

// CommandAttempt fires whenever a player submits command-prefixed chat.
type CommandAttempt struct {
	// Player stores the immutable command sender snapshot.
	Player sdkplayer.Player
	// Input stores the command text without the configured prefix.
	Input string
	// Root stores the first command token when one was supplied.
	Root string
}

// Name returns the stable command-attempt event identifier.
func (*CommandAttempt) Name() string { return CommandAttemptName }

// CloneEvent returns an isolated callback-owned notification.
func (event *CommandAttempt) CloneEvent() Event {
	cloned := *event
	return &cloned
}

// PlayerConnectedName identifies the authenticated-player notification.
const PlayerConnectedName = "player.connected"

// PlayerConnected fires after a player has an authenticated live session.
type PlayerConnected struct {
	// Player stores the immutable connected-player snapshot.
	Player sdkplayer.Player
}

// Name returns the stable player event identifier.
func (*PlayerConnected) Name() string { return PlayerConnectedName }

// CloneEvent returns an isolated callback-owned notification.
func (event *PlayerConnected) CloneEvent() Event {
	cloned := *event
	return &cloned
}
