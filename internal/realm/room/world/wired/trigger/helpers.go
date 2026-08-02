package trigger

import (
	"strings"
	"unicode/utf8"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// containsFold reports whether text contains a Unicode case-insensitive keyword without allocation.
func containsFold(value string, keyword string) bool {
	if keyword == "" {
		return true
	}
	if strings.Contains(value, keyword) {
		return true
	}
	for start := 0; start < len(value); {
		for end := start; end < len(value); {
			_, size := utf8.DecodeRuneInString(value[end:])
			end += size
			if strings.EqualFold(value[start:end], keyword) {
				return true
			}
		}
		_, size := utf8.DecodeRuneInString(value[start:])
		start += size
	}
	return false
}

// actorAllowed reports whether event actor kind satisfies the descriptor.
func actorAllowed(node *configuration.Node, actor ActorKind) bool {
	switch node.Descriptor.Actor {
	case registry.ActorPlayer:
		return actor == ActorPlayer
	case registry.ActorUnit:
		return actor == ActorPlayer || actor == ActorBot || actor == ActorPet
	case registry.ActorBot:
		return actor == ActorBot
	default:
		return true
	}
}

// botNameMatches compares an optional compiled bot name.
func botNameMatches(node *configuration.Node, event Event) bool {
	return node.Parameters.Name == "" || strings.EqualFold(node.Parameters.Name, event.Username)
}

// targetMatches applies Nitro's explicit ID, type, and context policies without allocation.
func targetMatches(node *configuration.Node, itemID int64, spriteID int32) bool {
	if node.SelectionMode == 0 {
		return false
	}
	for _, target := range node.Targets {
		if target.ItemID == itemID {
			return true
		}
		if node.SelectionMode >= 2 && spriteID > 0 && spriteID == target.SpriteID {
			return true
		}
	}
	return false
}

// clickedFurnitureMatches applies the official trigger, selected, and selector source modes.
func clickedFurnitureMatches(node *configuration.Node, event Event) bool {
	source := int32(0)
	if len(node.Parameters.Values) > 0 {
		source = node.Parameters.Values[0]
	}
	switch source {
	case 0:
		return node.ItemID == event.SourceItem
	case 100:
		return targetMatches(node, event.SourceItem, event.SourceSprite)
	case 200:
		return true
	default:
		return false
	}
}

// kindFor maps canonical trigger keys to typed room events.
func kindFor(key string) Kind {
	switch key {
	case "wf_trg_enter_room":
		return EnterRoom
	case "wf_trg_says_something":
		return Say
	case "wf_trg_walks_on_furni":
		return WalkOn
	case "wf_trg_walks_off_furni":
		return WalkOff
	case "wf_trg_state_changed":
		return StateChanged
	case "wf_trg_collision":
		return Collision
	case "wf_trg_periodically":
		return Periodic
	case "wf_trg_period_long":
		return PeriodicLong
	case "wf_trg_at_given_time":
		return AtTime
	case "wf_trg_at_time_long":
		return AtTimeLong
	case "wf_trg_game_starts":
		return GameStarted
	case "wf_trg_game_ends":
		return GameEnded
	case "wf_trg_score_achieved":
		return ScoreAchieved
	case "wf_trg_bot_reached_stf":
		return BotReachedFurniture
	case "wf_trg_bot_reached_avtr":
		return BotReachedAvatar
	case "wf_trg_game_team_win":
		return TeamWon
	case "wf_trg_game_team_lose":
		return TeamLost
	case "wf_trg_recv_signal":
		return ReceiveSignal
	case "wf_trg_leave_room":
		return LeaveRoom
	case "wf_trg_user_performs_action":
		return UserPerformsAction
	case "wf_trg_clock_counter":
		return ClockCounter
	case "wf_trg_var_changed":
		return VariableChanged
	case "wf_trg_user_clicks_furni":
		return FurnitureClicked
	case "wf_trg_user_clicks_tile":
		return FloorTileClicked
	case "wf_trg_user_clicks_user":
		return AvatarClicked
	default:
		return 0
	}
}

// Label returns one stable Creator Tools event label.
func Label(kind Kind) string {
	switch kind {
	case EnterRoom:
		return "ENTER_ROOM"
	case Say:
		return "SAY"
	case WalkOn:
		return "WALK_ON"
	case WalkOff:
		return "WALK_OFF"
	case StateChanged:
		return "STATE_CHANGED"
	case Collision:
		return "COLLISION"
	case Periodic:
		return "PERIODIC"
	case PeriodicLong:
		return "PERIODIC_LONG"
	case AtTime:
		return "AT_TIME"
	case AtTimeLong:
		return "AT_TIME_LONG"
	case GameStarted:
		return "GAME_STARTED"
	case GameEnded:
		return "GAME_ENDED"
	case ScoreAchieved:
		return "SCORE_ACHIEVED"
	case BotReachedFurniture:
		return "BOT_REACHED_FURNITURE"
	case BotReachedAvatar:
		return "BOT_REACHED_AVATAR"
	case TeamWon:
		return "TEAM_WON"
	case TeamLost:
		return "TEAM_LOST"
	case ReceiveSignal:
		return "RECEIVE_SIGNAL"
	case LeaveRoom:
		return "LEAVE_ROOM"
	case UserPerformsAction:
		return "USER_PERFORMS_ACTION"
	case ClockCounter:
		return "CLOCK_COUNTER"
	case VariableChanged:
		return "VARIABLE_CHANGED"
	case FurnitureClicked:
		return "FURNITURE_CLICKED"
	case FloorTileClicked:
		return "FLOOR_TILE_CLICKED"
	case AvatarClicked:
		return "AVATAR_CLICKED"
	default:
		return "UNKNOWN"
	}
}
