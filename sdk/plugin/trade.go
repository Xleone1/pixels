package plugin

// TradeParticipant stores one immutable live trade side.
type TradeParticipant struct {
	// PlayerID identifies the participant.
	PlayerID int64
	// UnitID identifies the participant's room unit.
	UnitID int64
	// Username stores the visible name snapshot.
	Username string
	// ItemIDs stores copied offered furniture identifiers.
	ItemIDs []int64
	// Accepted reports first-phase acceptance.
	Accepted bool
	// Confirmed reports final confirmation.
	Confirmed bool
}

// TradeSnapshot stores one immutable active trade.
type TradeSnapshot struct {
	// RoomID identifies the shared room.
	RoomID int64
	// First stores the initiating participant.
	First TradeParticipant
	// Second stores the invited participant.
	Second TradeParticipant
}

// TradeAccess exposes bounded active trade behavior.
type TradeAccess interface {
	// Active returns one player's active trade.
	Active(int64) (TradeSnapshot, bool)
	// ForceCancel closes one active trade with an audit reason.
	ForceCancel(int64, string) error
}
