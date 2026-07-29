package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

const (
	// PlayerProfileUpdateName identifies a cancellable public profile mutation.
	PlayerProfileUpdateName = "player.profile.update"
	// BotSpeechName identifies cancellable bot speech.
	BotSpeechName = "bot.speech"
	// GroupMembershipChangeName identifies a cancellable group membership mutation.
	GroupMembershipChangeName = "group.membership.change"
	// FriendRequestName identifies a cancellable friend request.
	FriendRequestName = "messenger.friend.request"
	// FriendAcceptName identifies a cancellable friend acceptance.
	FriendAcceptName = "messenger.friend.accept"
	// CraftingCraftName identifies a cancellable crafting transaction.
	CraftingCraftName = "crafting.craft"
)

// PlayerProfileUpdate fires before figure or motto persistence.
type PlayerProfileUpdate struct {
	// Player stores the immutable player snapshot.
	Player sdkplayer.Player
	// Motto stores the optional mutable motto.
	Motto *string
	// Figure stores the optional mutable figure string.
	Figure *string
	// Gender stores the optional mutable figure gender.
	Gender *string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable profile-update event identifier.
func (*PlayerProfileUpdate) Name() string { return PlayerProfileUpdateName }

// Cancelled reports whether the profile update was vetoed.
func (event *PlayerProfileUpdate) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the profile update is vetoed.
func (event *PlayerProfileUpdate) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned profile update.
func (event *PlayerProfileUpdate) Clone() Mutable {
	cloned := *event
	cloned.Motto = clonePointer(event.Motto)
	cloned.Figure = clonePointer(event.Figure)
	cloned.Gender = clonePointer(event.Gender)
	return &cloned
}

// Apply copies mutable profile state onto the original event.
func (event *PlayerProfileUpdate) Apply(original Mutable) {
	if target, ok := original.(*PlayerProfileUpdate); ok {
		target.Motto, target.Figure = clonePointer(event.Motto), clonePointer(event.Figure)
		target.Gender, target.cancelled = clonePointer(event.Gender), event.cancelled
	}
}

// BotSpeech fires after native filtering and before broadcast.
type BotSpeech struct {
	// BotID identifies the speaking bot.
	BotID int64
	// RoomID identifies the room receiving the speech.
	RoomID int64
	// Message stores the mutable filtered message.
	Message string
	// Scope identifies talk, shout, or whisper delivery.
	Scope string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable bot-speech event identifier.
func (*BotSpeech) Name() string { return BotSpeechName }

// Cancelled reports whether bot speech was vetoed.
func (event *BotSpeech) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether bot speech is vetoed.
func (event *BotSpeech) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned bot speech.
func (event *BotSpeech) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies mutable bot speech state onto the original event.
func (event *BotSpeech) Apply(original Mutable) {
	if target, ok := original.(*BotSpeech); ok {
		target.Message, target.cancelled = event.Message, event.cancelled
	}
}

// GroupMembershipChange fires before a join, add, or acceptance.
type GroupMembershipChange struct {
	// Action identifies join, add, or accept.
	Action string
	// Player stores the immutable affected player snapshot.
	Player sdkplayer.Player
	// GroupID identifies the target group.
	GroupID int64
	// Role identifies the requested role when applicable.
	Role string
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable group-membership event identifier.
func (*GroupMembershipChange) Name() string { return GroupMembershipChangeName }

// Cancelled reports whether the membership mutation was vetoed.
func (event *GroupMembershipChange) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the membership mutation is vetoed.
func (event *GroupMembershipChange) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned membership mutation.
func (event *GroupMembershipChange) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies the veto state onto the original event.
func (event *GroupMembershipChange) Apply(original Mutable) {
	if target, ok := original.(*GroupMembershipChange); ok {
		target.cancelled = event.cancelled
	}
}

// FriendRequest fires before a friend request is persisted.
type FriendRequest struct {
	// Sender stores the immutable requester snapshot.
	Sender sdkplayer.Player
	// Target stores the immutable recipient snapshot.
	Target sdkplayer.Player
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable friend-request event identifier.
func (*FriendRequest) Name() string { return FriendRequestName }

// Cancelled reports whether the request was vetoed.
func (event *FriendRequest) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the request is vetoed.
func (event *FriendRequest) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned friend request.
func (event *FriendRequest) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies the veto state onto the original event.
func (event *FriendRequest) Apply(original Mutable) {
	if target, ok := original.(*FriendRequest); ok {
		target.cancelled = event.cancelled
	}
}

// FriendAccept fires before a friendship is persisted.
type FriendAccept struct {
	// Actor stores the immutable accepting player snapshot.
	Actor sdkplayer.Player
	// Requester stores the immutable original requester snapshot.
	Requester sdkplayer.Player
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable friend-accept event identifier.
func (*FriendAccept) Name() string { return FriendAcceptName }

// Cancelled reports whether the acceptance was vetoed.
func (event *FriendAccept) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether the acceptance is vetoed.
func (event *FriendAccept) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned friend acceptance.
func (event *FriendAccept) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies the veto state onto the original event.
func (event *FriendAccept) Apply(original Mutable) {
	if target, ok := original.(*FriendAccept); ok {
		target.cancelled = event.cancelled
	}
}

// CraftingCraft fires before ingredients are consumed and a reward is granted.
type CraftingCraft struct {
	// Player stores the immutable crafter snapshot.
	Player sdkplayer.Player
	// RecipeID identifies the matched recipe.
	RecipeID int64
	// RewardDefinitionID stores the mutable reward definition.
	RewardDefinitionID int64
	// cancelled stores the current veto state.
	cancelled bool
}

// Name returns the stable crafting event identifier.
func (*CraftingCraft) Name() string { return CraftingCraftName }

// Cancelled reports whether crafting was vetoed.
func (event *CraftingCraft) Cancelled() bool { return event.cancelled }

// SetCancelled changes whether crafting is vetoed.
func (event *CraftingCraft) SetCancelled(value bool) { event.cancelled = value }

// Clone returns an isolated callback-owned crafting event.
func (event *CraftingCraft) Clone() Mutable {
	cloned := *event
	return &cloned
}

// Apply copies mutable crafting state onto the original event.
func (event *CraftingCraft) Apply(original Mutable) {
	if target, ok := original.(*CraftingCraft); ok {
		target.RewardDefinitionID, target.cancelled = event.RewardDefinitionID, event.cancelled
	}
}
