// Package event contains typed events exposed to dynamic plugins.
package event

import (
	"context"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
	sdkpriority "github.com/niflaot/pixels/sdk/priority"
)

// Event identifies one plugin-facing event.
type Event interface {
	// Name returns the stable event identifier.
	Name() string
}

// Cancellable describes an event whose default action may be vetoed.
type Cancellable interface {
	Event
	// Cancelled reports whether a listener vetoed the default action.
	Cancelled() bool
	// SetCancelled changes whether the default action is vetoed.
	SetCancelled(bool)
}

// Mutable describes an event whose callback-owned state can be committed safely.
type Mutable interface {
	Event
	// Clone returns an isolated copy for one plugin callback.
	Clone() Mutable
	// Apply copies this callback result onto the original event.
	Apply(Mutable)
}

// Listener reacts to one dispatched plugin-facing event.
type Listener func(context.Context, Event) error

// Published is an immutable post-commit realm notification.
type Published struct {
	// name stores the stable realm event identifier.
	name string
	// fields stores a normalized payload snapshot without internal realm types.
	fields map[string]any
}

// NewPublished creates one immutable post-commit event snapshot.
func NewPublished(name string, fields map[string]any) *Published {
	return &Published{name: name, fields: cloneFields(fields)}
}

// Name returns the stable realm event identifier.
func (event *Published) Name() string { return event.name }

// Fields returns a caller-owned copy of every normalized payload field.
func (event *Published) Fields() map[string]any { return cloneFields(event.fields) }

// Field returns one normalized payload field.
func (event *Published) Field(name string) (any, bool) {
	value, found := event.fields[name]
	return cloneValue(value), found
}

// CloneEvent returns one independent immutable notification.
func (event *Published) CloneEvent() Event {
	return NewPublished(event.name, event.fields)
}

// ListenerOptions configures event listener ordering and cancellation behavior.
type ListenerOptions struct {
	// Priority runs larger values before smaller values.
	Priority sdkpriority.Priority
	// IgnoreCancelled skips this listener after an earlier listener cancels.
	IgnoreCancelled bool
}

// cloneFields copies one normalized field map.
func cloneFields(fields map[string]any) map[string]any {
	cloned := make(map[string]any, len(fields))
	for name, value := range fields {
		cloned[name] = cloneValue(value)
	}
	return cloned
}

// cloneValue copies normalized composite values.
func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneFields(typed)
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneValue(item)
		}
		return cloned
	default:
		return typed
	}
}

// ChatSendName identifies the cancellable pre-delivery room chat event.
const ChatSendName = "chat.send"

// ChatSend fires before one sanitized room chat message is delivered.
type ChatSend struct {
	// Player stores the immutable speaker snapshot.
	Player sdkplayer.Player
	// RoomID identifies the room receiving the message.
	RoomID int64
	// Text stores the sanitized message and may be replaced by a listener.
	Text string
	// cancelled stores the current veto state.
	cancelled bool
}

// NewChatSend creates a cancellable room chat event.
func NewChatSend(player sdkplayer.Player, roomID int64, text string) *ChatSend {
	return &ChatSend{Player: player, RoomID: roomID, Text: text}
}

// Name returns the stable chat event identifier.
func (*ChatSend) Name() string { return ChatSendName }

// Cancelled reports whether room delivery was vetoed.
func (event *ChatSend) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether room delivery is vetoed.
func (event *ChatSend) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned chat event.
func (event *ChatSend) Clone() Mutable {
	cloned := NewChatSend(event.Player, event.RoomID, event.Text)
	cloned.SetCancelled(event.Cancelled())
	return cloned
}

// Apply copies this callback result onto the original chat event.
func (event *ChatSend) Apply(original Mutable) {
	target, ok := original.(*ChatSend)
	if !ok {
		return
	}
	target.Text = event.Text
	target.SetCancelled(event.Cancelled())
}

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
