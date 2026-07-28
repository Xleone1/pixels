package event

import (
	"time"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

const (
	// SanctionApplyName identifies a cancellable global sanction request.
	SanctionApplyName = "sanction.apply"
	// RoomModerationActionName identifies a cancellable room moderation request.
	RoomModerationActionName = "room.moderation.action"
)

// SanctionApply fires before any global sanction is persisted or applied.
type SanctionApply struct {
	// Actor stores the issuer or an empty snapshot for system issuance.
	Actor sdkplayer.Player
	// Target stores the sanction target.
	Target sdkplayer.Player
	// Kind identifies warn, mute, kick, ban, or trade lock.
	Kind string
	// Reason stores the mutable sanction explanation.
	Reason string
	// ExpiresAt optionally stores the mutable sanction expiration.
	ExpiresAt *time.Time
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable global sanction event identifier.
func (*SanctionApply) Name() string { return SanctionApplyName }

// Cancelled reports whether the sanction was vetoed.
func (event *SanctionApply) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the sanction is vetoed.
func (event *SanctionApply) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned sanction event.
func (event *SanctionApply) Clone() Mutable {
	cloned := *event
	cloned.ExpiresAt = clonePointer(event.ExpiresAt)
	return &cloned
}

// Apply copies this callback result onto the original sanction event.
func (event *SanctionApply) Apply(original Mutable) {
	target, ok := original.(*SanctionApply)
	if !ok {
		return
	}
	target.Reason = event.Reason
	target.ExpiresAt = clonePointer(event.ExpiresAt)
	target.SetCancelled(event.Cancelled())
}

// RoomModerationAction fires after native authorization and before mutation.
type RoomModerationAction struct {
	// Action identifies kick, mute, unmute, ban, or unban.
	Action string
	// RoomID identifies the affected room.
	RoomID int64
	// Actor stores the room moderator.
	Actor sdkplayer.Player
	// Target stores the affected player.
	Target sdkplayer.Player
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable room moderation event identifier.
func (*RoomModerationAction) Name() string { return RoomModerationActionName }

// Cancelled reports whether room moderation was vetoed.
func (event *RoomModerationAction) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether room moderation is vetoed.
func (event *RoomModerationAction) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned room moderation event.
func (event *RoomModerationAction) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies this callback result onto the original room moderation event.
func (event *RoomModerationAction) Apply(original Mutable) {
	target, ok := original.(*RoomModerationAction)
	if !ok {
		return
	}
	target.SetCancelled(event.Cancelled())
}

const (
	// TradeStartName identifies a cancellable trade opening request.
	TradeStartName = "trade.start"
	// TradeConfirmName identifies a cancellable final settlement request.
	TradeConfirmName = "trade.confirm"
	// TradeCancelName identifies a cancellable active trade closure request.
	TradeCancelName = "trade.cancel"
)

// TradeParticipant stores one immutable side of a live trade.
type TradeParticipant struct {
	// Player stores the participant snapshot.
	Player sdkplayer.Player
	// UnitID identifies the participant's room unit.
	UnitID int64
	// Items stores copied offered furniture identifiers.
	Items []int64
	// Accepted reports first-phase acceptance.
	Accepted bool
	// Confirmed reports final confirmation.
	Confirmed bool
}

// TradeStart fires before a new trade enters the live registry.
type TradeStart struct {
	// RoomID identifies the shared active room.
	RoomID int64
	// First stores the initiating participant.
	First TradeParticipant
	// Second stores the invited participant.
	Second TradeParticipant
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable trade start event identifier.
func (*TradeStart) Name() string { return TradeStartName }

// Cancelled reports whether opening the trade was vetoed.
func (event *TradeStart) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether opening the trade is vetoed.
func (event *TradeStart) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned trade start event.
func (event *TradeStart) Clone() Mutable { return cloneTradeStart(event) }

// Apply copies this callback result onto the original trade start event.
func (event *TradeStart) Apply(original Mutable) {
	target, ok := original.(*TradeStart)
	if ok {
		target.SetCancelled(event.Cancelled())
	}
}

// TradeConfirm fires after both sides confirm and before settlement.
type TradeConfirm struct {
	// RoomID identifies the shared active room.
	RoomID int64
	// First stores the first participant.
	First TradeParticipant
	// Second stores the second participant.
	Second TradeParticipant
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable trade confirmation event identifier.
func (*TradeConfirm) Name() string { return TradeConfirmName }

// Cancelled reports whether final settlement was vetoed.
func (event *TradeConfirm) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether final settlement is vetoed.
func (event *TradeConfirm) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned trade confirmation event.
func (event *TradeConfirm) Clone() Mutable {
	cloned := *event
	cloned.First.Items = append([]int64(nil), event.First.Items...)
	cloned.Second.Items = append([]int64(nil), event.Second.Items...)
	return &cloned
}

// Apply copies this callback result onto the original confirmation event.
func (event *TradeConfirm) Apply(original Mutable) {
	target, ok := original.(*TradeConfirm)
	if ok {
		target.SetCancelled(event.Cancelled())
	}
}

// TradeCancel fires before an active trade is closed.
type TradeCancel struct {
	// RoomID identifies the shared active room.
	RoomID int64
	// Actor stores the player whose action requested closure.
	Actor sdkplayer.Player
	// First stores the first participant.
	First TradeParticipant
	// Second stores the second participant.
	Second TradeParticipant
	// Reason stores the caller-provided cancellation reason.
	Reason string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable trade cancellation event identifier.
func (*TradeCancel) Name() string { return TradeCancelName }

// Cancelled reports whether closing the trade was vetoed.
func (event *TradeCancel) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether closing the trade is vetoed.
func (event *TradeCancel) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned trade cancellation event.
func (event *TradeCancel) Clone() Mutable {
	cloned := *event
	cloned.First.Items = append([]int64(nil), event.First.Items...)
	cloned.Second.Items = append([]int64(nil), event.Second.Items...)
	return &cloned
}

// Apply copies this callback result onto the original cancellation event.
func (event *TradeCancel) Apply(original Mutable) {
	target, ok := original.(*TradeCancel)
	if ok {
		target.SetCancelled(event.Cancelled())
	}
}

// cloneTradeStart copies one trade start event and its item slices.
func cloneTradeStart(event *TradeStart) *TradeStart {
	cloned := *event
	cloned.First.Items = append([]int64(nil), event.First.Items...)
	cloned.Second.Items = append([]int64(nil), event.Second.Items...)
	return &cloned
}
