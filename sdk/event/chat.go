package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

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
