// Package trigger matches compiled WIRED triggers against room events.
package trigger

import (
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// Kind classifies one triggerable room event.
type Kind uint8

const (
	// EnterRoom reports a completed player spawn.
	EnterRoom Kind = iota + 1
	// Say reports filtered player or bot speech.
	Say
	// WalkOn reports completed movement onto furniture.
	WalkOn
	// WalkOff reports completed movement away from furniture.
	WalkOff
	// StateChanged reports a committed furniture state change.
	StateChanged
	// Collision reports a movement collision.
	Collision
	// Periodic reports a short repeating timer deadline.
	Periodic
	// PeriodicLong reports a long repeating timer deadline.
	PeriodicLong
	// AtTime reports a short one-shot timer deadline.
	AtTime
	// AtTimeLong reports a long one-shot timer deadline.
	AtTimeLong
	// GameStarted reports a game lifecycle transition.
	GameStarted
	// GameEnded reports a game lifecycle transition.
	GameEnded
	// ScoreAchieved reports a score threshold crossing.
	ScoreAchieved
	// BotReachedFurniture reports bot arrival at a furniture target.
	BotReachedFurniture
	// BotReachedAvatar reports bot arrival at an avatar.
	BotReachedAvatar
	// TeamWon reports one winning team member.
	TeamWon
	// TeamLost reports one losing team member.
	TeamLost
	// ReceiveSignal reports an intra-room WIRED signal.
	ReceiveSignal
	// LeaveRoom reports a player leaving an active room.
	LeaveRoom
	// UserPerformsAction reports a wave, dance, posture, or gesture action.
	UserPerformsAction
	// ClockCounter reports a WIRED clock counter mutation.
	ClockCounter
	// VariableChanged reports a WIRED variable mutation.
	VariableChanged
)

// ActorKind classifies an event actor.
type ActorKind uint8

const (
	// ActorNone represents a system event.
	ActorNone ActorKind = iota
	// ActorPlayer represents a room player.
	ActorPlayer
	// ActorBot represents a room bot.
	ActorBot
	// ActorPet represents a room pet.
	ActorPet
)

// Event stores immutable room-event context.
type Event struct {
	// ID identifies the event inside one room lifecycle.
	ID uint64
	// Kind classifies the event.
	Kind Kind
	// RoomID identifies the containing room.
	RoomID int64
	// ActorKind classifies the actor.
	ActorKind ActorKind
	// ActorID identifies the runtime actor.
	ActorID int64
	// PlayerID identifies a stable player actor.
	PlayerID int64
	// Username stores the normalized player or bot name.
	Username string
	// SourceItem identifies contextual furniture.
	SourceItem int64
	// SourceSprite identifies contextual furniture type.
	SourceSprite int32
	// Message stores filtered deliverable speech.
	Message string
	// Score stores current game score.
	Score int64
	// PreviousScore stores score before the current mutation.
	PreviousScore int64
	// Team stores a game team from one through four.
	Team int32
	// Action stores the actor action identifier.
	Action int32
	// Direction stores the actor facing direction.
	Direction int32
	// Signal stores an intra-room signal name.
	Signal string
	// Counter stores the current WIRED clock counter value.
	Counter int64
	// PreviousCounter stores the prior WIRED clock counter value.
	PreviousCounter int64
	// VariableName stores the mutated variable definition name.
	VariableName string
	// VariableValue stores the current integer variable value.
	VariableValue int64
	// PreviousVariableValue stores the prior integer variable value.
	PreviousVariableValue int64
	// ReferenceRoomID identifies a remote room variable definition.
	ReferenceRoomID int64
	// ReferenceVariable stores the remote room variable name.
	ReferenceVariable string
	// ActorIDs stores actors carried by a WIRED signal.
	ActorIDs []int64
	// FurniTargets stores furniture carried by a WIRED signal.
	FurniTargets []record.Target
}

// Matcher matches one compiled trigger without allocation.
type Matcher struct{}

// New creates a stateless trigger matcher.
func New() Matcher { return Matcher{} }

// Match reports whether an event activates one trigger node.
func (Matcher) Match(node *configuration.Node, event Event) bool {
	if node == nil || node.RoomID != event.RoomID || !actorAllowed(node, event.ActorKind) {
		return false
	}
	expected := kindFor(node.Descriptor.Key)
	if expected == 0 || expected != event.Kind {
		return false
	}
	switch node.Descriptor.Key {
	case "wf_trg_says_something":
		return node.Parameters.Text != "" && containsFold(event.Message, node.Parameters.Text)
	case "wf_trg_enter_room":
		return node.Parameters.Text == "" || strings.EqualFold(node.Parameters.Text, event.Username)
	case "wf_trg_score_achieved":
		return len(node.Parameters.Values) > 0 && event.PreviousScore < int64(node.Parameters.Values[0]) && event.Score >= int64(node.Parameters.Values[0])
	case "wf_trg_clock_counter":
		return len(node.Parameters.Values) > 0 && event.PreviousCounter < int64(node.Parameters.Values[0]) && event.Counter >= int64(node.Parameters.Values[0])
	case "wf_trg_user_performs_action":
		return len(node.Parameters.Values) > 0 && event.Action == node.Parameters.Values[0]
	case "wf_trg_recv_signal":
		return node.Parameters.Text != "" && strings.EqualFold(node.Parameters.Text, event.Signal)
	case "wf_trg_var_changed":
		return node.Parameters.Text == "" || strings.EqualFold(node.Parameters.Text, event.VariableName)
	case "wf_trg_bot_reached_stf":
		return botNameMatches(node, event) && targetMatches(node, event.SourceItem, event.SourceSprite)
	case "wf_trg_bot_reached_avtr":
		return botNameMatches(node, event)
	case "wf_trg_walks_on_furni", "wf_trg_walks_off_furni", "wf_trg_state_changed":
		return targetMatches(node, event.SourceItem, event.SourceSprite)
	default:
		return true
	}
}
